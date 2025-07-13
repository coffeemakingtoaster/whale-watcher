package main

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/container"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout}).With().Caller().Logger()
	ociLocation := os.Args[1]
	c, _ := container.ContainerImageFromOCITar(ociLocation)
	fmt.Println(c.ToString())
	for i, layer := range c.Layers {
		fmt.Printf("%d - %s\n", i, layer.Command)
	}
	fmt.Println(c.Layers[0].Command)
}
