package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/thomasv314/helmui/pkg/watch"
)

var (
	chosenDriver string
	err          error
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	debug := flag.Bool("debug", false, "sets log level to debug")
	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Info().Msg("Debug logging is enabled")
	}

	if err != nil {
		panic(err.Error())
	}

	helmDriver := os.Getenv("HELM_DRIVER")
	if helmDriver == "" {
		helmDriver = watch.SecretStoreType
	}

	if helmDriver == watch.SecretStoreType || helmDriver == watch.ConfigMapStoreType {
		chosenDriver = helmDriver
	} else {
		panic(fmt.Errorf("helm driver not supported: %s", helmDriver))
	}

	log.Info().Str("release-driver", chosenDriver).Msg("Starting helmui")

	watch.Init(chosenDriver)

	stopCh := make(chan struct{})
	releaseWatcher := watch.NewReleaseWatcher()
	releaseWatcher.Run(stopCh)
	for {
		time.Sleep(time.Second)
	}
}
