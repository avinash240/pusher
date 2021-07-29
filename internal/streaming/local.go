package streaming

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// LocalAudio is a data structure that stores the path of the audio source(s).
type LocalAudio struct {
	FilePaths  []string
	StrictPath string
}

// StreamingData is a data structure that retains data bytes for pushing to
// endpoints.
type StreamingData struct {
	Bytes []byte
}

// NewLocalStream returns a pointer to an instance of LocalAudio with FilePaths
// translated for local audio data source(s).
func NewLocalStream(path string) (*LocalAudio, error) {
	var files []string
	p, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	fS, err := p.Stat()
	if err != nil {
		return nil, err
	}
	absP, _ := filepath.Abs(path)
	if fS.Mode().IsRegular() { //Is regular file
		files = append(files, path)
		pathSplit := strings.Split(absP, "/")
		pathSplit = pathSplit[:len(pathSplit)-1]
		absP = strings.Join(pathSplit, "/")
		return &LocalAudio{FilePaths: files, StrictPath: absP}, nil
	}
	filepath.Walk(path, func(spath string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			pathA := strings.Split(spath, "/")
			filename := pathA[len(pathA)-1]
			tar := strings.Join([]string{"/", filename}, "")
			files = append(files, tar)
		}
		return nil
	})

	return &LocalAudio{FilePaths: files, StrictPath: absP}, nil
}

// GetStream interfaces and instance of LocalAudio and returns a StreamingData
// channel. Byte channels contains streamed data from file. No paths found will
// return an error.
func (la *LocalAudio) GetStream() (sd chan StreamingData, e error) {
	if len(la.FilePaths) <= 0 {
		return nil, fmt.Errorf("no files found for the path provided")
	}
	sd = make(chan StreamingData)
	go streamDataToChannel(la.FilePaths, sd, 4096)
	return sd, nil
}

// streamDataToChannel takes a list of paths, a StreamingData channel, and a buf
// size for streaming. Data is streamed into channel from file in cS sized
// chunks. Channel closes after all files are read.
func streamDataToChannel(f []string, sd chan StreamingData, cS int) {
	buffer := make([]byte, cS)
	for _, p := range f {
		fh, err := os.Open(p)
		if err != nil {
			log.Fatalln(err)
		}
		for {
			n, err := fh.Read(buffer)
			chunk := make([]byte, n)
			copy(chunk, buffer)
			sd <- StreamingData{Bytes: chunk}
			if err != nil {
				if err == io.EOF {
					err = nil
				}
				break
			}
		}
		fh.Close()
	}
	close(sd)
}
