package main

import (
	"fmt"
	"log"

	"github.com/avinash240/pusher/internal/plugins"
	"github.com/avinash240/pusher/internal/streaming"
)

func main() {
	plugins, err := plugins.LoadPlugins("./config/plugins")
	if err != nil {
		log.Fatalln(err)
	}
	for _, p := range plugins {
		fmt.Printf("%+v", p)
	}

	/* Testing Local Streaming Code */
	fmt.Println("\n\n*** Testing Local Streaming ***")
	song, err := (&streaming.LocalAudio{}).New("file://localhost/tmp/data.txt")
	if err != nil {
		log.Fatalln(err)
	}
	c, err := song.GetStream()
	if err != nil {
		log.Fatalln(err)
	}

	for val := range c {
		log.Println(val)
	}
	fmt.Println("\n*****")
}
