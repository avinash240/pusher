package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	srv "github.com/avinash240/pusher/internal/server"
)

func TestLoadChromecastMiddleware(t *testing.T) {
	c := srv.NewHandler(false)
	if c == nil {
		t.Errorf("unable to load handler")
		t.FailNow()
	}
	go c.Serve("127.0.0.1:8081")
	time.Sleep(500 * time.Millisecond)
	resp, err := http.Get("http://localhost:8081/devices")
	if err != nil {
		t.Errorf(err.Error())
		t.FailNow()
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf(err.Error())
		t.FailNow()
	}
	type Data struct {
		Addr string `json:"addr"`
		Port int    `json:"port"`
		Name string `json:"DeviceName"`
		UUID string `json:"uuid"`
	}
	d := []Data{}
	json.Unmarshal(body, &d)
	log.Println(string(body))
	for k, v := range d {
		log.Println(k, v)
	}
	if len(d) > 0 && len(d[0].UUID) > 0 {
		resp, err := http.Get("http://localhost:8081/connect?uuid=" + d[0].UUID)
		if err != nil {
			t.Errorf(err.Error())
			t.FailNow()
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf(err.Error())
			t.FailNow()
		}
		if strings.Contains(string(body), "missing from query params") {
			t.Errorf(string(body))
			t.FailNow()
		}
	}
}
