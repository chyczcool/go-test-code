package main

import (
"fmt"
"sync"
"sync/atomic"
"time"
)

func main() {
var ch atomic.Value
ch.Store(make(chan string, 10))
var wg sync.WaitGroup
//comsumer
for i := 0; i < 5; i++ {
wg.Add(1)
go func(ch atomic.Value, i int) {
defer wg.Done()
fmt.Printf("comsumer-%d start.\n", i)
c := ch.Load()
channel, ok := c.(chan string)
if !ok {
fmt.Println("Invalid channel type")
return
}
for {
select {
case v, ok := <-channel:
if ok {
fmt.Printf("comsumer-%d print [%s]\n", i, v)
} else {
fmt.Printf("comsumer-%d exit.\n", i)
return
}
default:
time.Sleep(time.Second)
fmt.Printf("comsumer-%d sleep 1s.\n", i)
}
}
}(ch, i)
}

//producer
wg.Add(1)
go func(ch atomic.Value) {
c := ch.Load()
channel, ok := c.(chan string)
if !ok {
fmt.Println("Invalid channel type.")
return
}

defer func() {
fmt.Println("producer eixt.")
close(channel)
wg.Done()
}()

for i := 0; i < 100; i++ {
channel <- fmt.Sprintf("produce %d", i)
time.Sleep(250 * time.Millisecond)
fmt.Println("produce sleep 250ms")
}

}(ch)

wg.Wait()
}
