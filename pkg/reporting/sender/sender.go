package sender

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/loft-sh/magpie/pkg/events"
	"github.com/loft-sh/magpie/pkg/reporting"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/klog/v2"
)

type storeReader interface {
	ListAll(ctx context.Context) ([]events.KeyedEvent, error)
	GetGVK() schema.GroupVersionKind
}

type storeClearer interface {
	ClearDeletedResourceEvents(ctx context.Context) error
}

type Sender struct {
	URL     string
	Reader  storeReader
	Clearer storeClearer
	Client  *http.Client
}

func (s *Sender) Run(ctx context.Context) error {
	timer := time.Tick(1 * time.Minute)
	for {
		select {
		case <-timer:
			events, err := s.Reader.ListAll(ctx)
			if err != nil {
				klog.Errorf("failed to list all events: %v", err)
				continue
			}

			if len(events) == 0 {
				continue
			}

			eventsBundle := reporting.EventsBundle{
				GVK:    s.Reader.GetGVK(),
				Events: events,
			}

			eventsBytes, err := json.Marshal(eventsBundle)
			if err != nil {
				klog.Errorf("failed to marshal events: %v", err)
				continue
			}

			bytesReader := bytes.NewReader(eventsBytes)

			resp, err := s.Client.Post(s.URL, "application/json", bytesReader)
			if err != nil {
				klog.Errorf("failed to send events: %v", err)
				continue
			}

			if resp.StatusCode != http.StatusOK {
				klog.Errorf("failed to send events: %s", resp.Status)
				continue
			}
			respBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				klog.Errorf("failed to read response body: %v", err)
				continue
			}

			if len(respBody) != 0 {
				klog.Infof("events: %v", string(respBody))
			}

			// clear
			err = s.Clearer.ClearDeletedResourceEvents(ctx)
			if err != nil {
				klog.Errorf("failed to clear events: %v", err)
				continue
			}
		case <-ctx.Done():
			return nil
		}
	}
}
