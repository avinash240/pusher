package main

import (
	"log"
	"math"
	"strings"
	"testing"

	ls "github.com/avinash240/pusher/internal/streaming"
)

func TestLocalStream(t *testing.T) {
	strRp := 100
	log.Println(strings.Repeat("*", strRp))

	// Test against local file. Passes if file opened, streamed,
	// and len(data) > 0.
	localStream, err := ls.NewLocalStream("./test_data/thank_you.wav")
	log.Printf("* Test for file path: %v", localStream.FilePaths)
	if err != nil {
		t.Errorf("NewLocalStream() failed with issue:\n%+v", err)
		t.FailNow()
	}
	dataChannel, _ := localStream.GetStream()
	if dataChannel == nil {
		t.Errorf("Expected channel address; got nil instead.")
		t.FailNow()
	}
	log.Printf("*\t got StreamingData channel channel: %+v", dataChannel)
	for v := range dataChannel {
		if len(v.Bytes) <= 0 {
			t.Errorf("GetStream() reading data failed; no data read, or empty file.")
			t.FailNow()
		} else {
			end := math.Min(float64(len(v.Bytes)), 8)
			log.Printf("*\t  (%d,%0.0f) got data: [ %+s ... ]",
				len(v.Bytes),
				end,
				v.Bytes[:int(end)])
			break // Tests a single pass.
		}
	}
	log.Println(strings.Repeat("*", strRp))

	// Test against local files. Passes if more than one file is listed,
	// and len(data) > 0.
	path := "./test_data/"
	localStream, err = ls.NewLocalStream(path)
	log.Printf("* Test for directory path: %s", path)
	if err != nil {
		t.Errorf("NewLocalStream() failed with issue:\n%+v", err)
		t.FailNow()
	}
	if len(localStream.FilePaths) < 2 {
		t.Errorf("NewLocalStream() failed with issue: expected more than 2 files, got %d",
			len(localStream.FilePaths))
		t.FailNow()
	}
	log.Printf("*\t got files count: %d", len(localStream.FilePaths))
	log.Println(strings.Repeat("*", strRp))

	// Test against protected file. Passes if perrmission denied.
	path = "/etc/shadow"
	localStream, err = ls.NewLocalStream(path)
	log.Printf("* Test for file path: %s", path)
	if err == nil {
		t.Errorf("NewLocalStream() opened restricted file(s): %v", localStream.FilePaths)
		t.FailNow()
	}
	log.Printf("*    got expected error: %+v", err)
	log.Println(strings.Repeat("*", strRp))

	// Test against calling GetStream() without an instance of LocalAudio
	_, err = (&ls.LocalAudio{}).GetStream()
	log.Printf("* Test for bad method call on GetStream()")
	if err == nil {
		t.Errorf("GetStream() expected error; got nil")
		t.FailNow()
	} else {
		log.Printf("*\t got error as expected: \"%+s...\"", err.Error()[:len(err.Error())-20])
	}
	log.Println(strings.Repeat("*", strRp))

}
