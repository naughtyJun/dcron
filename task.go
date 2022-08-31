package dcron

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"math/rand"
	"time"
)

type (
	CronTask interface {
		Name() string
		Cron() string
		BeforeHook() error // init or status's judge in this
		Run()
	}
)

func RunOnce(task CronTask) {
	defer func() {
		if err := recover(); err != nil {
			logrus.WithField("err", err).
				Errorf("task[%s] panic", task.Name())
		}
	}()

	start := time.Now().Unix()
	logrus.Infof("task[%s] start", task.Name())

	if err := task.BeforeHook(); err != nil {
		logrus.WithField("err", err.Error()).
			Errorf("task[%s] %s error", task.Name(), "beforeHooks")
	} else {
		task.Run()
	}
	logrus.Infof("task[%s] end, use %d(second)", task.Name(), time.Now().Unix()-start)
}

func GenTaskId(prefix string) string {
	value := time.Now().Unix()
	rand.Seed(value)
	return fmt.Sprintf("task_id:%s-%d-%d", prefix, value, rand.Intn(1000))
}
