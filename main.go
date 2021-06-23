package main

import (
	"fmt"
	"log"

	"github.com/avinash240/pusher/internal/plugins"
)

func main() {
	plugins, err := plugins.LoadPlugins("./config/plugins")
	if err != nil {
		log.Fatalln(err)
	}
	for _, p := range plugins {
		fmt.Printf("%+v", p)
	}
}
