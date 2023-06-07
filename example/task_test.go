package example

import (
	"fmt"
	"testing"
	"time"

	"gitlab.bianjie.ai/spark/common-modules/dcron"
)

func TestDistributedTask(t *testing.T) {
	//集群节点数量
	nodeNum := 10
	for i := 0; i < nodeNum; i++ {
		//使用redis分布式锁，每次只有其中1个节点会执行任务
		d := dcron.NewDistributedTask(RClient, "spark", "console_server")
		//不使用分布式锁，10个节点都执行任务
		//d := dcron.NewDistributedTask(&dcron.WithoutLock{}, "spark", "console_server")
		d.RegisterTasks(
			&TestTask{
				node: i,
			},
		)
		d.Start()
	}

	select {}
}

//task skip
var _ dcron.CronTask = new(TestTask)

type TestTask struct {
	node int
}

func (t TestTask) Name() string {
	return "test_task"
}

func (t TestTask) Cron() string {
	return "*/10 * * * * *"
}

func (t TestTask) BeforeHook() error {
	return nil
}

func (t TestTask) Run() {
	fmt.Println(t.Name(), "node:", t.node, "begin:", time.Now().String()[:19])
	time.Sleep(15 * time.Second)
	fmt.Println(t.Name(), "node:", t.node, "end:", time.Now().String()[:19])
}

func (t TestTask) Mode() dcron.Mode {
	return dcron.ModeSkipIfStillRunning
}
