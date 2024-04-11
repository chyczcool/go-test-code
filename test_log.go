package main

import (
	"sync"
	"time"

	"github.com/tanpopoycz/go-test-code/module/log"
)

func main() {
	log.Init()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		for i := 0; i < 100000; i++ {
			log.Debug().Int("i", i).Msg("some debug msg")
			time.Sleep(time.Millisecond * 400)
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		for i := 0; i < 100000; i++ {
			log.Info().Int("i", i).Msg("some info msg")
			time.Sleep(time.Millisecond * 2)
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		for i := 0; i < 10000; i++ {
			log.Error().Int("i", i).Msg("some info msg")
			time.Sleep(time.Millisecond * 30)
		}
		wg.Done()
	}()

	wg.Wait()
}
