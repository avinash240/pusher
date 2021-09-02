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
	if err := json.NewEncoder(w).Encode(chttp.ConnectResponse{DeviceUUID: deviceUUID}); err != nil {
		log.Printf("error encoding json: %v", err)
		httpError(w, fmt.Errorf("unable to json encode devices: %v", err))
		return
	}

}

func httpError(w http.ResponseWriter, err error) {
	w.Header().Add("Content-Type", "text/plain")
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func httpValidationError(w http.ResponseWriter, msg string) {
	http.Error(w, msg, http.StatusBadRequest)
}
