package example

import (
	"fmt"
	"testing"
	"time"

	"gitlab.bianjie.ai/spark/common-modules/dcron"
)

func TestTaskMode(t *testing.T) {
	t.Log("start at: ", time.Now().String()[:19])

	d := dcron.NewDistributedTask(&dcron.WithoutLock{}, "spark", "console_server")
	//如果用了分布式锁，default模式和skip的结果是一致的
	//因为前面的任务没执行完，后面的任务执行了也获取不到锁，还是相当于skip了
	//但是性能不如skip模式，因为skip模式是直接程序内判断，不需要额外的网络IO（redis lock）
	//d := dcron.NewDistributedTask(RClient, "spark", "console_server")
	d.RegisterTasks(
		&TestTaskDefault{},
		&TestTaskSkip{},
		&TestTaskDelay{},
	)
	d.Start()

	select {}
}

//task default
var _ dcron.CronTask = new(TestTaskDefault)

var count int

type TestTaskDefault struct {
}

func (t TestTaskDefault) Mode() dcron.Mode {
	return dcron.ModeDefault
}

func (t TestTaskDefault) Name() string {
	return "test_task_default"
}

func (t TestTaskDefault) Cron() string {
	return "*/10 * * * * *"
}

func (t TestTaskDefault) BeforeHook() error {
	return nil
}

func (t TestTaskDefault) Run() {
	count++
	fmt.Println(t.Name(), "begin: ", time.Now().String()[:19])
	//if count <= 5 {
	//	time.Sleep(9 * time.Second)
	//} else {
	//	time.Sleep(1 * time.Second)
	//}
	time.Sleep(15 * time.Second)
	fmt.Println(t.Name(), "end: ", time.Now().String()[:19])
}

//task skip
var _ dcron.CronTask = new(TestTaskSkip)

type TestTaskSkip struct {
}

func (t TestTaskSkip) Name() string {
	return "test_task_skip"
}

func (t TestTaskSkip) Cron() string {
	return "*/10 * * * * *"
}

func (t TestTaskSkip) BeforeHook() error {
	return nil
}

func (t TestTaskSkip) Run() {
	fmt.Println(t.Name(), "begin: ", time.Now().String()[:19])
	time.Sleep(15 * time.Second)
	fmt.Println(t.Name(), "end: ", time.Now().String()[:19])
}

func (t TestTaskSkip) Mode() dcron.Mode {
	return dcron.ModeSkipIfStillRunning
}

//task delay
var _ dcron.CronTask = new(TestTaskDelay)

type TestTaskDelay struct {
}

func (t TestTaskDelay) Name() string {
	return "test_task_delay"
}

func (t TestTaskDelay) Cron() string {
	return "*/10 * * * * *"
}

func (t TestTaskDelay) BeforeHook() error {
	return nil
}

func (t TestTaskDelay) Run() {
	fmt.Println(t.Name(), "begin: ", time.Now().String()[:19])
	time.Sleep(15 * time.Second)
	fmt.Println(t.Name(), "end: ", time.Now().String()[:19])
}

func (t TestTaskDelay) Mode() dcron.Mode {
	return dcron.ModeDelayIfStillRunning
}
