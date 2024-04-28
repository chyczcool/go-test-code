package testsync

import (
	"fmt"
	"time"
)

var ch = make(chan string, 10)

func TestMutiWriteChan() {
	for i := 0; i < 3; i++ {
		go func(i int) {
			for {
				s := fmt.Sprintf("this is %d", i)
				ch <- s
				time.Sleep(time.Second)
			}
		}(i)
	}

	for {
		select {
		case s := <-ch:
			fmt.Println(s)
		default:
			time.Sleep(time.Millisecond)
		}
	}
}
