package transfer

import (
	"fmt"
	"log"
	"testing"
)

func Test_client(t *testing.T) {
	//c := NewClient(addressFileC)
	//c.Dial()
	//c.Download("google-chrome-stable_current_amd64.deb","google-chrome-stable_current_amd64.deb.1")
	dispatcher := NewDispatcher(MAXWORKERS)
	go func() {
		dispatcher.Run()
	}()

	JobQueue = make(chan Job, 1)
	isJobFinish = make(chan bool, 100)

	c := NewClient(AddressFileC)
	c.Dial()
	stat, err := c.Stat("data.log")
	if err != nil {
		t.Errorf("get error:%v", err)
	}
	blocks := int(stat.Size / BLOCK_SIZE)
	if stat.Size%BLOCK_SIZE != 0 {
		blocks += 1
	}

	downLoadBlocks := 0

	for {
		name := fmt.Sprintf("data.log.%d", downLoadBlocks)
		log.Println(name)
		select {
		case JobQueue <- Job{downLoadBlocks, 0, "data.log", name}:
			log.Println("send job to Queue")
		}
		downLoadBlocks = downLoadBlocks + 1
		log.Println("downLoadBlock is:", downLoadBlocks, blocks)
		if downLoadBlocks >= blocks {
			break
		}

	}
	tasksNum := 0
	for {
		select {
		case <-isJobFinish:
			log.Println("taskNum is finished:", tasksNum)
			tasksNum = tasksNum + 1
		}
		if tasksNum >= blocks {
			log.Println("==============================================", tasksNum)
			break
		}
	}
}
