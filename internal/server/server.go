package server

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	ls "github.com/avinash240/pusher/internal/streaming"
)

// getLocalAddress returns IP address for eth0 or wlan0, or error if no matching
// interface is found. Also returns error if no IPv4 address is assigned.
func getLocalAddress() (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var iface net.Interface
	for _, eth := range ifaces {
		ethName := strings.ToLower(eth.Name)
		if strings.Contains(ethName, "eth0") ||
			strings.Contains(ethName, "wlan0") ||
			strings.Contains(ethName, "wi-fi") ||
			strings.Contains(ethName, "ethernet") ||
			strings.Contains(ethName, "wireless") {
			iface = eth
			break
		}
	}
	if iface.Name == "" {
		return nil, fmt.Errorf("no eth0 or wlan0 interface for use")
	}
	var IPaddr net.IP
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}
	for _, add := range addrs {
		ipA, _, err := net.ParseCIDR(add.String())
		if err != nil {
			return nil, err
		}
		if ipA.To4() != nil {
			IPaddr = ipA
			return IPaddr, nil
		}
	}
	return nil, fmt.Errorf("no usable addresses found")
}

// sendMsg prints out logging/debugging messages to standard out
func sendMsg(s string) {
	log.Println(s)
}

// streamItem is data structure for loaded media items
type streamItem struct {
	filename    string
	contentType string
	contentURL  string
	transcode   bool
}

// load takes http.Response, http.Request from http handler, and the server
// address and port number; returns an array of streaming items or error.
// load is used by webserver as a http handle function. load walks the directory
// specified to the webserver in the media_file parameter, and load assets into
// data structure for streaming. load must be ran first or server will not serve
// content.
func load(w http.ResponseWriter, r *http.Request, address net.IP, port int) ([]streamItem, error) {
	target := r.URL.Query().Get("target")
	//transcode := r.URL.Query().Get("live_streaming")
	//TODO: something with live_streaming transcoding with ffmeg?
	if target == "" {
		msg := fmt.Sprintf("missing target parameter in %s", r.URL)
		sendMsg(msg)
		http.Error(w, msg, 400)
		return nil, fmt.Errorf(msg)
	}
	if strings.Contains(target, "://") { //Do we expect remote URL here?
		msg := fmt.Sprintf("Remote URL not supported: %s", target)
		sendMsg(msg)
		http.Error(w, msg, 400)
		return nil, fmt.Errorf(msg)
	}

	var streamItems []streamItem
	assets, err := ls.NewLocalStream(target)
	if err != nil {
		sendMsg(err.Error())
		http.Error(w, err.Error(), 400)
		return nil, err
	}
	streamItems = make([]streamItem, len(assets.FilePaths))
	for i, item := range assets.FilePaths {
		filename := filepath.Base(item)
		p := filepath.Join(assets.StrictPath, filename)
		streamItems[i].filename = p
		ext := filepath.Ext(filename)
		if ext == "" {
			streamItems[i].contentType = "Unknown"
		} else {
			streamItems[i].contentType = ext
		}

		streamItems[i].transcode = false //TODO: something about transcoding
		url := fmt.Sprintf(
			"http://%s:%d?media_file=%s&live_streaming=%t",
			address.String(),
			port,
			streamItems[i].filename,
			streamItems[i].transcode,
		)
		streamItems[i].contentURL = url
	}
	msg := fmt.Sprintf("loadded assets in %s", target)
	sendMsg(msg)
	fmt.Fprint(w, "Loaded assets.")
	return streamItems, nil
}

// mediaServer is used by webserver as a http handle function. mediaServer will
// list for a URL formate to serve content from loaded media. load() needs to be
// ran first
func mediaServer(w http.ResponseWriter, r *http.Request, sI []streamItem, loaded bool) error {
	if !loaded {
		msg := "no media loaded"
		sendMsg(msg)
		http.Error(w, msg, 400)
		return fmt.Errorf(msg)
	}
	// URL Format: media_file=%s&live_streaming=%t
	qs := strings.Split(r.URL.RawQuery, "&")
	mf := r.URL.Query().Get("media_file")
	//ls, err := strconv.ParseBool(r.URL.Query().Get("live_streaming")) // live stream transcoding?
	//if err != nil {
	// 	http.Error(w,err.Err(),400)
	// 	return err
	// }
	if mf == "" {
		msg := fmt.Sprintf("media_file parameter not found in %s", qs)
		sendMsg(msg)
		http.Error(w, msg, 400)
		return fmt.Errorf(msg)
	}
	contentAvailable := false
	var idx []int

	for i, v := range sI {
		if strings.Contains(v.filename, mf) {
			contentAvailable = true
			idx = append(idx, i)
		}
	}
	if len(idx) >= 2 {
		msgI := fmt.Sprintf("'%s' matches multiple loaded; try again", mf)
		sendMsg(msgI)
		matches := []string{msgI, "\nid,url"}
		for _, i := range idx {
			s := fmt.Sprintf("%d,%s", i, sI[i].contentURL)
			matches = append(matches, s)
		}
		msg := strings.Join(matches, "\n")
		http.Error(w, msg, 400)
		return fmt.Errorf(msgI)
	}
	if !contentAvailable {
		msg := fmt.Sprintf("%s not found in loaded media", mf)
		sendMsg(msg)
		http.Error(w, msg, 400)
		return fmt.Errorf(msg)
	} else {
		mf = sI[idx[0]].filename
	}

	//TODO: something for live_streaming parameter? ffmpeg?

	http.ServeFile(w, r, mf)
	msg := fmt.Sprintf("served file: %s", mf)
	sendMsg(msg)
	return nil
}

// contentQuery is used by webserver to display a list of loaded assets or error
// if no loaded content.
func contentQuery(w http.ResponseWriter, r *http.Request, sI []streamItem, loaded bool) {
	if !loaded {
		msg := fmt.Sprintln("no media loaded")
		sendMsg(msg)
		http.Error(w, msg, 400)
		return
	}
	queryGet := r.URL.Query().Get("id")
	if queryGet != "" {
		item, err := strconv.Atoi(queryGet)
		if err != nil {
			sendMsg(err.Error())
			http.Error(w, err.Error(), 400)
			return
		}
		if item >= len(sI) {
			msg := fmt.Sprintf("'%d' is outside of list index", item)
			sendMsg(msg)
			http.Error(w, msg, 400)
			return
		} else {
			msg := fmt.Sprintf("sent details for %s", sI[item].filename)
			sendMsg(msg)
			v := fmt.Sprintf("{\n  \"filename\": \"%s\",\n  \"contentType\": \"%s\",\n  \"contentURL\": \"%s\",\n  \"transcode\": %t\n}\n",
				sI[item].filename,
				sI[item].contentType,
				sI[item].contentURL,
				sI[item].transcode)
			fmt.Fprintf(w, "%s", v)
			return
		}
	}
	fmt.Fprint(w, "id,url\n")
	for i, v := range sI {
		msg := fmt.Sprintf("%d,%s\n", i, v.contentURL)
		fmt.Fprintf(w, "%s", msg)
	}
	sendMsg("Outputed list of assets")
}

// NewLocalServer starts a webserver on port 9002 and all available address. If
// port not available server will exit.
func NewLocalServer() {
	port := 9002
	loaded := false
	var sI []streamItem
	address, err := getLocalAddress()
	if err != nil {
		log.Fatalln(err)
	}

	msg := fmt.Sprintf("Listening on 0.0.0.0:%d", port)
	sendMsg(msg)

	http.HandleFunc("/load", func(w http.ResponseWriter, r *http.Request) {
		sI, err = load(w, r, address, port)
		if err == nil {
			loaded = true
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := mediaServer(w, r, sI, loaded)
		if err != nil {
			return
		}
	})

	http.HandleFunc("/content", func(w http.ResponseWriter, r *http.Request) {
		contentQuery(w, r, sI, loaded)
	})

	portStr := strconv.Itoa(port)
	listenAddr := strings.Join([]string{":", portStr}, "")
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
