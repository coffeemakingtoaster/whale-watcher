package main

import (
	"os"

	"github.com/coffeemakingtoaster/whale-watcher/internal/command"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout}).With().Logger()
	command.Run(os.Args[1:])
}
