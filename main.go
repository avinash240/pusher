package main

import (
	"fmt"
	"log"

	"github.com/avinash240/pusher/internal/plugins"
	strm "github.com/avinash240/pusher/internal/streaming"
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
	song, err := strm.NewLocalStream("internal/tests/test_data/thank_you.wav")
	if err != nil {
		log.Fatalln(err)
	}
	c, err := song.GetStream()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("\n\n%T <- %v\n", c, c)
}
