package main

import (
	"log"
	"math"
	"strings"
	"testing"

	strm "github.com/avinash240/pusher/internal/streaming"
)

func TestLocalStream(t *testing.T) {
	strRp := 50
	log.Println(strings.Repeat("*", strRp))

	// Test against local file. Passes if file opened, streamed,
	// and len(data) > 0.
	localStream, err := strm.NewLocalStream("./test_data/")
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
			log.Printf("\t  (%d,%0.0f) got data: [ %+s ... ]",
				len(v.Bytes),
				end,
				v.Bytes[:int(end)])
			break
		}
	}
	log.Println(strings.Repeat("*", strRp))

	// Test against protected file. Passes if perrmission denied.
	path := "/etc/shadow"
	localStream, err = strm.NewLocalStream(path)
	log.Printf("* Test for file path: %s", path)
	if err == nil {
		t.Errorf("NewLocalStream() opened restricted file(s): %v", localStream.FilePaths)
		t.FailNow()
	}
	log.Printf("*    got expected error: %+v", err)
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
