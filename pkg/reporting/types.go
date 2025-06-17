package reporting

import (
	"github.com/loft-sh/magpie/pkg/eventstores"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type EventsBundle struct {
	GVK    schema.GroupVersionKind  `json:"gvk"`
	Events []eventstores.KeyedEvent `json:"events"`
}
