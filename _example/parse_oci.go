package main

import (
	"fmt"
	"io"
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
	index := len(c.Layers) - 1
	for index >= 0 {
		fmt.Println(c.Layers[index].FileSystem.Ls("/etc/apt/"))
		ok, deletion := c.Layers[index].FileSystem.HasFile("/etc/apt/sources.list.d/debian.sources")
		if ok {
			if deletion {
				fmt.Print("deleted")
				return
			}
			fmt.Print(index)
		}
		index--
	}
	fmt.Print("done")
}
