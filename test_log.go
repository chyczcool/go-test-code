package main

import (
	"github.com/tanpopoycz/go-test-code/module/log"
)

func main() {
	log.Init()
	log.Info().Msg("some msg.")

}
