package streaming

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

// LocalAudio is a data structure that stores the URI of file and Transport scheme (usually file://)
type LocalAudio struct {
	URI       url.URL
	Transport http.Transport
}

type StreamingData struct {
	Bytes []byte
}

func (*LocalAudio) New(path string) (*LocalAudio, error) {
	uri, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	t := &http.Transport{}
	t.RegisterProtocol("file", http.NewFileTransport((http.Dir("/"))))
	return &LocalAudio{URI: *uri, Transport: *t}, nil
}

func (la *LocalAudio) GetStream() (sd chan []byte, e error) {
	sd = make(chan []byte)
	fileReader, filzeSize, err := getReader(la.URI, la.Transport)
	if err != nil {
		return nil, err
	}

	go pullData(fileReader, sd, make([]byte, filzeSize))
	return sd, e
}

func getReader(u url.URL, t http.Transport) (io.Reader, int, error) {
	c := &http.Client{}
	c.Transport = &t
	resp, err := c.Get(u.String())
	if err != nil {
		return nil, 0, err
	}
	fileSize, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return nil, 0, err
	}
	return resp.Body, fileSize, nil
}

func pullData(r io.Reader, s chan []byte, buf []byte) {
	n, err := r.Read(buf)
	if err != nil {
		log.Fatalln(err)
	}
	if n > 0 {
		s <- buf
	}
	close(s)
}
