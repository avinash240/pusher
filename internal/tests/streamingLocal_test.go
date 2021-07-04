package main

import (
	"go/build"
	"log"
	"os"
	"strings"
	"testing"

	strm "github.com/avinash240/pusher/internal/streaming"
)

func TestLocalStream(t *testing.T) {
	strRp := 50
	log.Println(strings.Repeat("*", strRp))
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}
	//Test against local file. Passes if 200 OK and len(data) > 0.
	path := strings.Join([]string{"file://localhost", gopath,
		"/src/pusher/internal/tests/test_data/thank_you.wav"}, "")
	file, err := (&strm.LocalAudio{}).New(path)
	log.Printf("* Test for file path: %s", file.URI.Path)
	if err != nil {
		t.Errorf("New() failed with issue:\n%+v", err)
		t.FailNow()
	}

	data, err := file.GetStream()
	if err != nil {
		t.Errorf("GetStream() failed with issue:\n%+v", err)
		t.FailNow()
	} else {
		log.Println("*\t got expected 200 OK")
		if data == nil {
			t.Errorf("Expected channel address; got nil instead.")
			t.FailNow()
		}
		log.Printf("*\t got channel: %+v", data)
		for v := range data {
			if len(v) <= 0 {
				t.Errorf("GetStream() reading data failed; no data read, or empty file.")
				t.FailNow()
			} else {
				log.Printf("\t got data: [ %+s ... ]", v[:8])
			}
		}
	}
	log.Println(strings.Repeat("*", strRp))

	// Test against protected file. Passes if 403 Forbidden is reflected.
	file, err = (&strm.LocalAudio{}).New("file://localhost/etc/shadow")
	log.Printf("* Test for file path: %s", file.URI.Path)
	if err != nil {
		t.Errorf("New() failed with issue:\n%+v", err)
		t.FailNow()
	}
	_, err = file.GetStream()
	if err != nil && strings.Contains(err.Error(), "403") == false {
		t.Errorf("GetStream() failed with issue:\n%+v", err)
		t.FailNow()
	} else if err != nil {
		log.Println("*\t got expected 403 Forbidden")
	} else {
		t.Errorf("Error 403 expected; got data instead.")
		t.FailNow()
	}
	log.Println(strings.Repeat("*", strRp))

	// Test against calling GetStream() without an instance of LocalAudio
	_, err = (&strm.LocalAudio{}).GetStream()
	log.Printf("* Test for bad method call on GetStream()")
	if err == nil {
		t.Errorf("GetStream() expected error; got nil")
		t.FailNow()
	} else {
		log.Printf("*\t got error as expected: \"%+s...\"", err.Error()[:len(err.Error())-20])
	}
	log.Println(strings.Repeat("*", strRp))

}
