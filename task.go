package dcron

import (
	"fmt"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	ModeDefault             Mode = iota //每到执行时间直接执行，不管前面任务执行是否完成
	ModeSkipIfStillRunning              //到了执行时间，前面任务执行未完成，则直接跳过，本次不执行，等待下次执行时间
	ModeDelayIfStillRunning             //到了执行时间，前面任务执行未完成，则等待其完成后立即执行(可能造成任务堆积，不建议使用)
)

type (
	Mode int

	CronTask interface {
		Name() string
		Cron() string
		BeforeHook() error // init or status's judge in this
		Run()
		Mode() Mode
	}
)

func RunOnce(task CronTask) {
	logger := logrus.WithField("task", task.Name())

	defer func() {
		if r := recover(); r != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
			logger.WithError(err).Errorf("panic, stack:%v", string(buf))
		}
	}()

	start := time.Now().Unix()
	logger.Debug("start")

	if err := task.BeforeHook(); err != nil {
		logger.WithField("method", "BeforeHook").WithError(err).Error()
	} else {
		task.Run()
	}
	logger.Infof("end, use %d(second)", time.Now().Unix()-start)
}
