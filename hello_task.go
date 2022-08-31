package dcron

import (
	"fmt"
	"time"
)

type HelloTask struct {
}

func (h *HelloTask) Name() string {
	return "hello"
}

func (h *HelloTask) BeforeHooks() error {
	return nil
}

func (h *HelloTask) AfterHooks() error {
	time.Sleep(3 * time.Second)
	fmt.Println("ccc")
	return nil
}

func (h *HelloTask) Cron() string {
	return "*/30 * * * * *"
}

func (h *HelloTask) Run() {
	fmt.Println("hello world do .... wait sleep")
	panic("panic this hello")
	//time.Sleep(5 * time.Second)
	fmt.Println("hello world done")
}
