package dcron

import (
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"time"
)

type DistributedTask struct {
	l          LockHasExpired
	expiration time.Duration
	tasks      []CronTask
}

const defaultExpiration = 30 * time.Second

func NewDistributedTask(l LockHasExpired) *DistributedTask {
	return &DistributedTask{l: l}
}

func (d *DistributedTask) Expiration() time.Duration {
	if d.expiration <= 10 {
		return defaultExpiration
	}
	return d.expiration
}

func (d *DistributedTask) SetExpiration(expiration time.Duration) {
	d.expiration = expiration
}

func (d *DistributedTask) RegisterTasks(task ...CronTask) {
	d.tasks = append(d.tasks, task...)
}

func (d *DistributedTask) Start() {
	if len(d.tasks) == 0 {
		return
	}
	if d.l == nil {
		logrus.Warn("you need implements LockHasExpired interface")
		d.l = new(WithoutLock)
	}

	c := cron.New(cron.WithSeconds())
	for _, v := range d.tasks {
		t := v
		_, err := c.AddFunc(t.Cron(), func() {
			d.RunOnceWithLock(t)
		})
		if err != nil {
			logrus.WithField("task", t.Name()).
				WithField("err", err.Error()).
				Fatal("add cron job err", err)
		}
	}
	c.Start()
}

func (d *DistributedTask) RunOnceWithLock(task CronTask) {
	logger := logrus.WithField("key", task.Name())
	value := GenTaskId(task.Name())
	if err := d.l.Lock(task.Name(), value, d.Expiration()); err != nil {
		logger.
			WithField("err", err.Error()).
			Errorf("redis lock failed")

		return
	}

	stop := make(chan struct{})
	t := time.NewTicker(10 * time.Second)
	go func() {
		for {
			select {
			case <-t.C:
				d.ReNewExpiration(task.Name(), value)
			case <-stop:
				logger.Debug("expired")
				return
			}
		}
	}()

	RunOnce(task)
	stop <- struct{}{}
	if _, err := d.l.UnLock(task.Name(), value); err != nil {
		logger.
			WithField("err", err.Error()).
			Error("unlock failed")
	}
}

func (d *DistributedTask) ReNewExpiration(key string, value interface{}) {
	logger := logrus.WithField("key", key)
	ttl, err := d.l.TTL(key)
	if err != nil {
		logger.Error("get ttl failed")
		return
	}
	if ttl == 0 {
		return
	}

	if ttl <= d.Expiration()/3 {
		if _, err := d.l.Expire(key, value, d.Expiration()); err != nil {
			logger.
				WithField("err", err.Error()).
				Error("expired failed")
		}
		logger.Debug("renew")
	}
}
