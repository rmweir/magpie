package sender

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/loft-sh/magpie/pkg/eventstores"
	"google.golang.org/appengine/log"
	"k8s.io/apimachinery/pkg/util/json"
)

type storeReader interface {
	ListAll(ctx context.Context) ([]eventstores.KeyedEvent, error)
	ClearEvents(ids ...string) error
}
type Sender struct {
	URL   string
	Store storeReader
}

func (s *Sender) Run(ctx context.Context) error {
	timer := time.Tick(1 * time.Minute)
	for {
		select {
		case <-timer:
			events, err := s.Store.ListAll(ctx)
			if err != nil {
				log.Errorf(ctx, "failed to list all events: %v", err)
				continue
			}

			eventsBytes, err := json.Marshal(events)
			if err != nil {
				log.Errorf(ctx, "failed to marshal events: %v", err)
				continue
			}
			bytesReader := bytes.NewReader(eventsBytes)
			resp, err := http.Post(s.URL, "application/json", bytesReader)
			if err != nil {
				log.Errorf(ctx, "failed to send events: %v", err)
				continue
			}
			respBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Errorf(ctx, "failed to read response body: %v", err)
				continue
			}
			if len(respBody) != 0 {
				log.Infof(ctx, "events: %v", string(respBody))
			}
		case <-ctx.Done():
			return nil
		}
	}
}
