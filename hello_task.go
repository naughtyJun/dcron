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

func (h *HelloTask) BeforeHook() error {
	return nil
}

func (h *HelloTask) Cron() string {
	return "*/12 * * * * *"
}

func (h *HelloTask) Run() {
	fmt.Println("hello world do .... wait sleep")
	time.Sleep(30 * time.Second)
	fmt.Println("hello world done")
}
