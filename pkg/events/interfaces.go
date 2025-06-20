package events

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type EventStore interface {
	ListAll(ctx context.Context) ([]KeyedEvent, error)
	List(ctx context.Context, key ResourceKey) ([]Event, error)
	Add(ctx context.Context, event ...KeyedEvent) error
	GetResourceKeyFromUnstructured(obj unstructured.Unstructured) ResourceKey
	ClearDeletedResourceEvents(ctx context.Context) error
}
