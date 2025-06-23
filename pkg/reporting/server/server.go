package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/loft-sh/magpie/pkg/reporting"
	"k8s.io/klog/v2"
)

const (
	disconnectThreshold = time.Duration(10 * time.Minute)
	EventCollectPath    = "/logs/collect"
	HealthCheckPath     = "/health"
)

type disconnectCallback func(string) error

type CollectorServer struct {
	*http.ServeMux
	dirPath            string
	peer               lru.Cache[string, time.Time]
	disconnectCallback disconnectCallback
}

var _ http.Handler = (*CollectorServer)(nil)

func NewCollectorServer(filePath string, disconnectCallback func(string) error) *CollectorServer {
	server := &CollectorServer{
		ServeMux: http.NewServeMux(),
		dirPath:  filePath,
	}

	server.ServeMux.Handle(EventCollectPath, http.HandlerFunc(server.collectLogs))
	if disconnectCallback != nil {
		server.disconnectCallback = disconnectCallback
		server.ServeMux.Handle(HealthCheckPath, http.HandlerFunc(server.healthPing))
		go func() {
			server.checkDisconnects()
		}()
	}

	return server
}

func (c *CollectorServer) collectLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	receivedLogs := reporting.EventsBundle{}
	logBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(logBytes, &receivedLogs)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = os.WriteFile(c.filepath(receivedLogs, time.Now()), logBytes, 0644)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	fmt.Print(receivedLogs.Events)
}

func (c *CollectorServer) healthPing(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	healthCheck := reporting.HealthPing{}
	logBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(logBytes, &healthCheck)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_ = c.peer.Add(healthCheck.Peer, time.Now())
}

func (c *CollectorServer) filepath(l reporting.EventsBundle, writeTime time.Time) string {
	if l.GVK.Group == "" {
		return fmt.Sprintf("%s/%s-%s-%s", c.dirPath, l.GVK.Kind, l.GVK.Version, writeTime.Format(time.RFC3339))
	}
	return fmt.Sprintf("%s/%s-%s-%s-%s", c.dirPath, l.GVK.Group, l.GVK.Kind, l.GVK.Version, writeTime.Format(time.RFC3339))
}

func (c *CollectorServer) checkDisconnects() {
	ticker := time.NewTicker(disconnectThreshold).C
	for {
		select {
		case <-ticker:
			for _, peer := range c.peer.Keys() {
				lastSeen, found := c.peer.Get(peer)
				if !found || time.Now().Sub(lastSeen) < disconnectThreshold {
					continue
				}

				err := c.disconnectCallback(peer)
				if err != nil {
					klog.Errorf("disconnect callback error for peer [%s]: %v", peer, err)
					continue
				}
				c.peer.Remove(peer)
			}
		}
	}
}
