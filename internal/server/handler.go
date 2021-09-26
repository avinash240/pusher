package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	application "github.com/avinash240/pusher/internal/server/application"
	chttp "github.com/avinash240/pusher/internal/server/chttp"
	dns "github.com/avinash240/pusher/internal/server/dns"
)

// Application Handler
type Handler struct {
	mu      sync.Mutex
	apps    map[string]*application.Application
	mux     *http.ServeMux
	verbose bool
}

// Device info data structure
type device struct {
	Addr string `json:"addr"`
	Port int    `json:"port"`

	Name string `json:"name"`
	Host string `json:"host"`

	UUID       string            `json:"uuid"`
	Device     string            `json:'device_type"`
	Status     string            `json:"status"`
	DeviceName string            `json"device_name"`
	InfoFields map[string]string `json"info_fields"`
}

func NewHandler(verbose bool) *Handler {
	handler := &Handler{
		verbose: verbose,
		apps:    map[string]*application.Application{},
		mux:     http.NewServeMux(),
		mu:      sync.Mutex{},
	}
	handler.registerHandlers()
	return handler
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) Serve(addr string) error {
	log.Printf("starting http server on %s", addr)
	return http.ListenAndServe(addr, h)
}

func (h *Handler) registerHandlers() {
	/*
		GET /devices
		POST /connect?uuid=<device_uuid>&addr=<device_addr>&port=<device_port>
		POST /disconnect?uuid=<device_uuid>
		POST /disconnect-all
		POST /status?uuid=<device_uuid>
		POST /pause?uuid=<device_uuid>
		POST /unpause?uuid=<device_uuid>
		POST /mute?uuid=<device_uuid>
		POST /unmute?uuid=<device_uuid>
		POST /stop?uuid=<device_uuid>
		GET /volume?uuid=<device_uuid>
		POST /volume?uuid=<device_uuid>&volume=<float>
		POST /rewind?uuid=<device_uuid>&seconds=<int>
		POST /seek?uuid=<device_uuid>&seconds=<int>
		POST /seek-to?uuid=<device_uuid>&seconds=<float>
		POST /load?uuid=<device_uuid>&path=<filepath_or_url>&content_type=<string>
	*/

	h.mux.HandleFunc("/devices", h.listDevices)
	h.mux.HandleFunc("/connect", h.connect)
	h.mux.HandleFunc("/disconnect", h.disconnect)
	// h.mux.HandleFunc("/disconnect-all", h.disconnectAll)
	// h.mux.HandleFunc("/status", h.status)
	// h.mux.HandleFunc("/pause", h.pause)
	// h.mux.HandleFunc("/unpause", h.unpause)
	// h.mux.HandleFunc("/mute", h.mute)
	// h.mux.HandleFunc("/unmute", h.unmute)
	// h.mux.HandleFunc("/stop", h.stop)
	// h.mux.HandleFunc("/volume", h.volume)
	// h.mux.HandleFunc("/rewind", h.rewind)
	// h.mux.HandleFunc("/seek", h.seek)
	// h.mux.HandleFunc("/seek-to", h.seekTo)
	h.mux.HandleFunc("/load", h.load)
}

func (h *Handler) app(uuid string) (*application.Application, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	app, ok := h.apps[uuid]
	return app, ok
}

func (h *Handler) discoverDnsEntries(ctx context.Context, iface string, waitq string) (devices []device) {
	wait := 3
	if n, err := strconv.Atoi(waitq); err == nil {
		wait = n
	}

	devices = []device{}
	var interf *net.Interface
	if iface != "" {
		var err error
		interf, err = net.InterfaceByName(iface)
		if err != nil {
			log.Printf("error discovering entries: %v", err)
			return
		}
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(wait)*time.Second)
	defer cancel()

	devicesChan, err := dns.DiscoverCastDNSEntries(ctx, interf)
	if err != nil {
		log.Printf("error discovering entries: %v", err)
		return
	}
	for d := range devicesChan {
		devices = append(devices, device{
			Addr:       d.AddrV4.String(),
			Port:       d.Port,
			Name:       d.Name,
			Host:       d.Host,
			UUID:       d.UUID,
			Device:     d.Device,
			Status:     d.Status,
			DeviceName: d.DeviceName,
			InfoFields: d.InfoFields,
		})
	}

	return
}

func (h *Handler) listDevices(w http.ResponseWriter, r *http.Request) {
	log.Println("Listing Chromecast Devices")

	q := r.URL.Query()
	iface := q.Get("interface")
	wait := q.Get("wait")

	devices := h.discoverDnsEntries(context.Background(), iface, wait)
	log.Printf("found %d devices", len(devices))

	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(devices); err != nil {
		log.Printf("error encoding json: %v", err)
		httpError(w, fmt.Errorf("unable to json encode devices: %v", err))
		return
	}
}

func (h *Handler) connect(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	deviceUUID := q.Get("uuid")
	if deviceUUID == "" {
		http.Error(w, "missing 'uuid' in query paramater", http.StatusBadRequest)
		return
	}

	_, ok := h.app(deviceUUID)
	if ok {
		http.Error(w, "device uuid is already connected", http.StatusBadRequest)
		return
	}

	deviceAddr := q.Get("addr")
	devicePort := q.Get("port")
	iface := q.Get("interface")
	wait := q.Get("wait")

	if deviceAddr == "" || devicePort == "" {
		log.Printf("device addr and/or port are missing, trying to lookup address for uuid %q", deviceUUID)

		devices := h.discoverDnsEntries(context.Background(), iface, wait)
		for _, device := range devices {
			if device.UUID == deviceUUID {
				deviceAddr = device.Addr
				devicePort = strconv.Itoa(device.Port)
			}
		}
	}

	if deviceAddr == "" || devicePort == "" {
		http.Error(w, "'port' and 'addr' missing from query params and uuid device lookup returned no results", http.StatusBadRequest)
		return
	}

	log.Printf("connecting to addr=%s port=%s...", deviceAddr, devicePort)

	devicePortI, err := strconv.Atoi(devicePort)
	if err != nil {
		log.Printf("device port %q is not a number: %v", devicePort, err)
		http.Error(w, "'port' is not a number", http.StatusBadRequest)
		return
	}

	applicationOptions := []application.ApplicationOption{
		application.WithDebug(h.verbose),
		application.WithCacheDisabled(true),
	}

	app := application.NewApplication(applicationOptions...)
	if err := app.Start(deviceAddr, devicePortI); err != nil {
		log.Printf("unable to start application: %v", err)
		httpError(w, fmt.Errorf("unable to start application: %v", err))
		return
	}
	h.mu.Lock()
	h.apps[deviceUUID] = app
	h.mu.Unlock()

	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(chttp.ConnectResponse{DeviceUUID: deviceUUID, Connected: true}); err != nil {
		log.Printf("error encoding json: %v", err)
		httpError(w, fmt.Errorf("unable to json encode devices: %v", err))
		return
	}

}

func (h *Handler) disconnect(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	deviceUUID := q.Get("uuid")
	if deviceUUID == "" {
		httpValidationError(w, "missing 'uuid' in query paramater")
		return
	}

	log.Printf("disconnecting device %s", deviceUUID)

	app, ok := h.app(deviceUUID)
	if !ok {
		httpValidationError(w, "device uuid is not connected")
		return
	}

	stopMedia := q.Get("stop") == "true"
	if err := app.Close(stopMedia); err != nil {
		log.Printf("unable to close application: %v", err)
	}

	h.mu.Lock()
	delete(h.apps, deviceUUID)
	h.mu.Unlock()
	fmt.Fprintf(w, "Disconnected from %v\n", deviceUUID)
}

func (h *Handler) load(w http.ResponseWriter, r *http.Request) {
	app, found := h.appForRequest(w, r)
	if !found {
		return
	}

	log.Println("loading media for device")

	q := r.URL.Query()
	path := q.Get("path")
	if path == "" {
		httpValidationError(w, "missing 'path' in query paramater")
		return
	}

	contentType := q.Get("content_type")

	if err := app.Load(path, contentType, true, true, true); err != nil {
		log.Printf("unable to load media for device: %v", err)
		httpError(w, fmt.Errorf("unable to load media for device: %w", err))
		return
	}
}

func (h *Handler) appForRequest(w http.ResponseWriter, r *http.Request) (*application.Application, bool) {
	q := r.URL.Query()

	deviceUUID := q.Get("uuid")
	if deviceUUID == "" {
		httpValidationError(w, "missing 'uuid' in query params")
		return nil, false
	}

	app, ok := h.app(deviceUUID)
	if !ok {
		httpValidationError(w, "device uuid is not connected")
		return nil, false
	}

	if err := app.Update(); err != nil {
		return nil, false
	}

	return app, true
}

func httpError(w http.ResponseWriter, err error) {
	w.Header().Add("Content-Type", "text/plain")
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func httpValidationError(w http.ResponseWriter, msg string) {
	http.Error(w, msg, http.StatusBadRequest)
}
