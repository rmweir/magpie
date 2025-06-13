package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

type CollectorServer struct {
	*http.ServeMux
}

type Logs struct {
	GVK    schema.GroupVersionKind `json:"gvk"`
	Events map[string]interface{}  `json:"events"`
}

var _ http.Handler = (*CollectorServer)(nil)

func NewCollectorServer() *CollectorServer {
	server := &CollectorServer{
		ServeMux: http.NewServeMux(),
	}
	http.HandleFunc("/logs/collect", server.ServeMux.ServeHTTP)
	return server
}

func (c *CollectorServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	receivedLogs := Logs{}
	err := json.NewDecoder(r.Body).Decode(&receivedLogs)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Print(receivedLogs.Events)
}
