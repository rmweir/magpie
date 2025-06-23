package reporting

import (
	"github.com/loft-sh/magpie/pkg/events"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type EventsBundle struct {
	GVK    schema.GroupVersionKind `json:"gvk"`
	Events []events.KeyedEvent     `json:"events"`
}

type HealthPing struct {
	Peer string `json:"peer"`
}
