package main

import (
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	srv "github.com/avinash240/pusher/internal/server"
)

func TestServer(t *testing.T) {
	strRp := 100
	log.Println(strings.Repeat("*", strRp))
	go srv.NewLocalServer()
	log.Println("* Initializing server")

	time.Sleep(1 * time.Second) // wait for web server to load
	resp, err := http.Get("http://localhost:9002/")
	if err != nil {
		t.Error(err)
		t.FailNow()
	} else {
		log.Println("*  server started")
	}
	if resp.StatusCode != 400 {
		t.Errorf("expected status 400 got %d", resp.StatusCode)
		t.FailNow()
	}
	resp, err = http.Get("http://100.115.92.202:9002/loadStream?target=./test_data/")
	if err != nil {
		t.Error(err)
		t.FailNow()
	} else {
		log.Println("*  media loaded")
	}
	if resp.StatusCode == 200 {
		log.Println("*  server loaded with assests")
	}

	log.Println(strings.Repeat("*", strRp))
	log.Println("* Requesting Media")
	resp, err = http.Get("http://localhost:9002/?media_file=ascii")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if resp.StatusCode != 400 {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
		t.FailNow()
	} else {
		log.Println("*  passed check on multiple file match for: ascii")
	}

	log.Println(strings.Repeat("*", strRp))
	log.Println("* Requesting Media")
	resp, err = http.Get("http://localhost:9002/?media_file=a_ascii.txt")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	s := resp.ContentLength
	buf := make([]byte, s)
	_, _ = resp.Body.Read(buf)
	if len(buf) < 1000 {
		t.Errorf("expected 1000 bytes of data, got %d", len(buf))
		t.FailNow()
	}
	log.Printf("*  got data %+s...", buf[:30])
	log.Println(strings.Repeat("*", strRp))
}
