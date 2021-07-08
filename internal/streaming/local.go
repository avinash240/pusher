package streaming

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

// LocalAudio is a data structure that stores the URI of file and Transport
// scheme (usually file://).
type LocalAudio struct {
	URI       url.URL
	Transport http.Transport
}

// StreamingData is a data structure that retains data bytes for pushing to
// endpoints.
type StreamingData struct {
	Bytes []byte
}

// New returns a pointer to an instance of LocalAudio with path translated
// for local audio data.
func NewLocalStream(path string) (*LocalAudio, error) {
	uri, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	t := &http.Transport{}
	t.RegisterProtocol("file", http.NewFileTransport((http.Dir("/"))))
	return &LocalAudio{URI: *uri, Transport: *t}, nil
}

// GetStream interfaces and instance of LocalAudio and returns a byte channel or
// err. Byte channels contains streamed data.
// NOTES - Return streamingdata type from channel
func (la *LocalAudio) GetStream() (streamingData chan []byte, e error) {
	streamingData = make(chan []byte)
	if len(la.URI.Path) <= 0 {
		return nil, fmt.Errorf(
			"no path provided. This needs LocalAudio.New(path) first; got '%+v'",
			la.URI.Path)
	}
	// NOTES - No longer needed os.Open() with give you something with the
	// io.Reader interface
	fileReader, filzeSize, err := getReader(la.URI, la.Transport)
	if err != nil {
		return nil, err
	}

	// NOTES - move make call into pulldata
	go pullData(fileReader, streamingData, make([]byte, filzeSize))
	return streamingData, e
}

// getReader is a private function that takes a path and http transport type and
// returns an io.Reader, size of file, or error.
func getReader(u url.URL, t http.Transport) (io.Reader, int, error) {
	c := &http.Client{}
	c.Transport = &t
	resp, err := c.Get(u.String())
	if resp.StatusCode > 399 || err != nil {
		if resp.StatusCode > 399 {
			err = fmt.Errorf("expected status 200 or redirection 30x. Got '%s'",
				resp.Status)
		}
		return nil, 0, err
	}
	fileSize, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return nil, 0, err
	}
	return resp.Body, fileSize, nil
}

// pullData is a private function that takes an io.Reader and channel and data
// buffer and writes data to the channel.
func pullData(r io.Reader, s chan []byte, buf []byte) {
	// NOTES - move this to io.ReadAll
	start := 0
	l := make([]byte, 3000)
	for {
		n, err := r.Read(l[start:])
		if err != nil {
			log.Fatalln(err)
		}
		start += n
		s <- buf
	}
	defer close(s)
}
