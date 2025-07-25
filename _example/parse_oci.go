package main

import (
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/container"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout}).With().Caller().Logger()
	ociLocation := os.Args[1]
	c, _ := container.ContainerImageFromOCITar(ociLocation)
	fmt.Println(c.ToString())
	for i, layer := range c.Layers {
		fmt.Printf("%d - %s\n", i, layer.Command)

		if ok, _ := layer.FileSystem.HasFile("/app/test.txt"); ok {
			f, err := layer.FileSystem.Open("/app/test.txt")
			if err != nil {
				panic(err)
			}
			data, err := io.ReadAll(f)
			if err != nil {
				panic(err)
			}
			fmt.Printf("%s\n", string(data))
		}
	}
	fmt.Print("extracting")
	err := c.ExtractToDir("./.extract")
	if err != nil {
		log.Error().Err(err).Send()
	}
}
