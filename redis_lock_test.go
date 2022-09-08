package dcron

import (
	"testing"
	"time"
)

func TestExpire(t *testing.T) {
	Init()
	err := RClient.Lock("fffff", 10000, 60*time.Second)
	if err != nil {
		t.Fatal(err.Error())
	}

	res, err := RClient.Expire("fffff", 10000, 600*time.Second)
	if err != nil {
		t.Fatal(err.Error())
	}
	if res.(int64) == 0 {
		t.Fatal("expire failed maybe the execute result is 0")
	}
}

func TestHelloTask(t *testing.T) {
	Init()
	d := NewDistributedTask(RClient)
	d.RegisterTasks(&HelloTask{})
	d.Start()
	select {}
}
