package events

import (
	"fmt"
	"k8s.io/apimachinery/pkg/types"
)

type ResourceKey struct {
	types.NamespacedName
	types.UID
}

func (r ResourceKey) String() string {
	return fmt.Sprintf("%s_%s_%s", r.Namespace, r.Name, r.UID)
}
