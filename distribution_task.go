package dcron

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

type DistributedTask struct {
	l          LockHasExpired
	expiration time.Duration
	tasks      []CronTask
	nameSpace  string
	serverName string
}

const (
	defaultExpiration   = 30 * time.Second
	reNewTickerDuration = 10 * time.Second
)

func NewDistributedTask(l LockHasExpired, nameSpace string, serverName string) *DistributedTask {
	return &DistributedTask{
		l:          l,
		nameSpace:  nameSpace,
		serverName: serverName,
	}
}

func (d *DistributedTask) Expiration() time.Duration {
	if d.expiration <= defaultExpiration {
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
	//c := cron.New(cron.WithSeconds(), cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
	//c := cron.New(cron.WithSeconds(), cron.WithChain(cron.DelayIfStillRunning(cron.DefaultLogger)))

	for _, v := range d.tasks {
		t := v
		entry := logrus.WithField("task", t.Name())
		log := newLogger(entry)
		var jobWrapper cron.JobWrapper
		switch t.Mode() {
		case ModeSkipIfStillRunning:
			jobWrapper = cron.SkipIfStillRunning(log)
		case ModeDelayIfStillRunning:
			jobWrapper = cron.DelayIfStillRunning(log)
		default:
			jobWrapper = func(job cron.Job) cron.Job { return job }
		}
		_, err := c.AddJob(t.Cron(), cron.NewChain(jobWrapper).Then(cron.FuncJob(func() {
			d.RunOnceWithLock(t)
		})))

		if err != nil {
			entry.WithError(err).Fatal("add cron job err")
		}
	}
	c.Start()
}

func (d *DistributedTask) RunOnceWithLock(task CronTask) {
	logger := logrus.WithField("key", task.Name())
	key := d.genLockKey(task.Name())
	value := d.genLockValue(task.Name())
	ok, err := d.l.Lock(key, value, d.Expiration())
	if err != nil {
		logger.WithError(err).Errorf("lock failed")
		return
	}
	if !ok {
		logger.Debug("not get lock")
		return
	}
	stop := make(chan struct{})
	t := time.NewTicker(reNewTickerDuration)
	defer t.Stop()

	go func() {
		for {
			select {
			case <-t.C:
				d.ReNewExpiration(key, value)
			case <-stop:
				logger.Debug("stop")
				return
			}
		}
	}()

	RunOnce(task)
	stop <- struct{}{}
	if _, err := d.l.UnLock(key, value); err != nil {
		logger.WithError(err).Error("unlock failed")
	}
}

func (d *DistributedTask) ReNewExpiration(key string, value interface{}) {
	logger := logrus.WithField("key", key)
	ttl, err := d.l.TTL(key)
	if err != nil {
		logger.WithError(err).Error("get ttl failed")
		return
	}
	if ttl == 0 {
		return
	}

	if ttl <= d.Expiration()/3 {
		if _, err := d.l.ExpireWithVal(key, value, d.Expiration()); err != nil {
			logger.WithError(err).Error("expired failed")
		}
		logger.Debug("renew")
	}
}

func (d *DistributedTask) genLockKey(taskName string) string {
	return fmt.Sprintf("%s:%s:lock:%s", d.nameSpace, d.serverName, taskName)
}

func (d *DistributedTask) genLockValue(taskName string) string {
	value := time.Now().UnixNano()
	hostname, _ := os.Hostname()
	rand.NewSource(value)
	return fmt.Sprintf("task_id:%s-%s-%d-%d", taskName, hostname, value, rand.Intn(100))
}
