package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/coffeemakingtoaster/whale-watcher/internal/command"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout}).With().Caller().Logger()
	command.Run(os.Args[1:])
}
