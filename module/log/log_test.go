package main

import (
	"flag"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// setup would normally be an init() function, however, there seems
// to be something awry with the testing framework when we set the
// global Logger from an init()
func setup() {

	zerolog.TimeFieldFormat = time.DateTime

	// zerolog.TimestampFunc = func() time.Time {
	// 	return time.Date(2008, 1, 8, 17, 5, 05, 0, time.UTC)
	// }
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
}

func main() {
	setup()
	debug := flag.Bool("debug", false, "sets log level to debug")

	flag.Parse()

	// Default level for this example is info, unless debug flag is present
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log.Debug().Msg("This message appears only when log level set to Debug")
	log.Info().Msg("This message appears when log level set to Debug or Info")
	log.Info().Msg("This message appears when log level set to Debug or Info")

	if e := log.Debug(); e.Enabled() {
		// Compute log output only if enabled.
		value := "bar"
		e.Str("foo", value).Msg("some debug message")
	}

}
