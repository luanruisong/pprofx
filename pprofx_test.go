package pprofx

import (
	"fmt"
	"testing"
	"time"
)

func TestRecordCpuProfile(t *testing.T) {

	AutoDuration(time.Second * 5)
	for {
		time.Sleep(time.Second)
		fmt.Println(time.Now().UnixMilli())
	}

}
