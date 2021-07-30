package streaming

import (
	"os"
	"path/filepath"
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
		return &LocalAudio{FilePaths: files, StrictPath: absP}, nil
	}
	filepath.Walk(path, func(spath string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			filename := filepath.Base(spath)
			p := filepath.Join(absP, filename)
			files = append(files, p)
		}
		return nil
	})
	return &LocalAudio{FilePaths: files, StrictPath: absP}, nil
}
