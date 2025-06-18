package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/loft-sh/magpie/pkg/reporting"
)

type CollectorServer struct {
	*http.ServeMux
	dirPath string
}

var _ http.Handler = (*CollectorServer)(nil)

func NewCollectorServer(filePath string) *CollectorServer {
	server := &CollectorServer{
		ServeMux: http.NewServeMux(),
		dirPath:  filePath,
	}
	http.HandleFunc("/logs/collect", server.ServeMux.ServeHTTP)
	return server
}

func (c *CollectorServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	err = os.WriteFile(c.Filepath(receivedLogs, time.Now()), logBytes, 0644)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	fmt.Print(receivedLogs.Events)
}

func (c *CollectorServer) Filepath(l reporting.EventsBundle, writeTime time.Time) string {
	if l.GVK.Group == "" {
		return fmt.Sprintf("%s/%s-%s-%s", c.dirPath, l.GVK.Kind, l.GVK.Version, writeTime.Format(time.RFC3339))
	}
	return fmt.Sprintf("%s/%s-%s-%s-%s", c.dirPath, l.GVK.Group, l.GVK.Kind, l.GVK.Version, writeTime.Format(time.RFC3339))
}
