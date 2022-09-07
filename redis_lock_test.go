package dcron

import (
	"fmt"
	"testing"
	"time"
)

func TestExpire(t *testing.T) {
	Init()
	err := RClient.Lock("fffff", 10000, 600*time.Second)
	if err != nil {
		fmt.Println(err.Error())
	}

	_, err = RClient.Expire("fffff", 1000, 1000*time.Second)
	if err != nil {
		fmt.Println(err.Error())
	}
}
