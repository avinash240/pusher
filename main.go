package main

import (
	"fmt"
	"log"

	// "github.com/avinash240/pusher/internal/plugins"
	srv "github.com/avinash240/pusher/internal/server"
)

func main() {
	// plugins, err := plugins.LoadPlugins("./config/plugins")
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// for _, p := range plugins {
	// 	fmt.Printf("%+v", p)
	// }

	// /* Testing Server Code*/
	go srv.NewLocalServer()

	log.Println("")
	/* Testing Chromecast Connect */
	c := srv.NewHandler(false)
	fmt.Printf("c: %v\n", c)

	s := c.Serve("127.0.0.1:8080")
	fmt.Printf("s: %v\n", s)
}
