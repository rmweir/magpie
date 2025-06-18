package events

import (
	"time"

	"k8s.io/apimachinery/pkg/types"
)

const (
	_ eventType = iota
	Create
	InferredCreate
	Delete
	InferredDelete
)

type eventType int

type Event struct {
	Obj       map[string]interface{} `json:"obj,omitempty"`
	EventType eventType              `json:"event_type,omitempty"`
	ObjUID    types.UID              `json:"obj_uid,omitempty"`
	Time      time.Time              `json:"time,omitempty"`
	EventID   string                 `json:"event_id,omitempty"`
}
