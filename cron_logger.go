package dcron

import (
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

var _ cron.Logger = new(cronLogger)

type cronLogger struct {
	entry *logrus.Entry
}

func newLogger(entry *logrus.Entry) cron.Logger {
	return cronLogger{entry: entry}
}

func (cl cronLogger) Info(msg string, keysAndValues ...interface{}) {
	cl.entry.WithField("kvs", keysAndValues).Info(msg)
}

func (cl cronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	cl.entry.WithError(err).WithField("kvs", keysAndValues).Error(msg)
}
