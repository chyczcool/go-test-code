package log

import (
	"sync"
	"time"
)

func Test_log() {
	Init()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		for i := 0; i < 100000; i++ {
			Debug().Int("i", i).Msg("some debug msg")
			time.Sleep(time.Millisecond * 400)
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		for i := 0; i < 100000; i++ {
			Info().Int("i", i).Msg("some info msg")
			time.Sleep(time.Millisecond * 200)
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		for i := 0; i < 10000; i++ {
			Warn().Int("i", i).Msg("some warn msg")
			time.Sleep(time.Millisecond * 400)
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		for i := 0; i < 10000; i++ {
			Error().Int("i", i).Msg("some info msg")
			time.Sleep(time.Millisecond * 600)
		}
		wg.Done()
	}()

	wg.Wait()
}
