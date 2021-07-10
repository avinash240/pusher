package streaming

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LocalAudio is a data structure that stores the path of the audio source, the
// file handle when opened, and the size of the file.
type LocalAudio struct {
	Location string
	File     os.File
	Size     int64
}

// StreamingData is a data structure that retains data bytes for pushing to
// endpoints.
type StreamingData struct {
	Bytes []byte
}

// NewLocalStream returns a pointer to an instance of LocalAudio with path translated
// for local audio data.
func NewLocalStream(path string) (*LocalAudio, error) {
	fh, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	fp, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	fs, err := fh.Stat()
	if err != nil {
		return nil, err
	}
	return &LocalAudio{Location: fp, File: *fh, Size: fs.Size()}, nil
}

// GetStream interfaces and instance of LocalAudio and returns a StreamingData
// channel. Byte channels contains streamed data from file. If no location found
// error will return.
func (la *LocalAudio) GetStream() (sd chan StreamingData, e error) {
	if la.Location == "" {
		return nil, fmt.Errorf("failed file opening, or no filepath specified;" +
			" see NewLocalAudio()")
	}
	sd = make(chan StreamingData)
	go streamDataToChannel(la.File, sd, 4096)
	return sd, nil
}

// streamDataToChannel is takes an os.File, a StreamingData channel, and a size
// to make starting buffer. Data is streamed into channel from file. Derived
// from io.ReadAll() source.
func streamDataToChannel(f os.File, sd chan StreamingData, cS int) {
	buffer := make([]byte, 0, cS)
	for {
		if len(buffer) == cap(buffer) {
			buffer = append(buffer, 0)[:len(buffer)]
		}
		n, err := f.Read(buffer[len(buffer):cap(buffer)])
		sd <- StreamingData{Bytes: buffer[len(buffer):cap(buffer)]}
		buffer = buffer[:len(buffer)+n]
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			close(sd)
			return
		}
	}
}
