package entity

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID        string    `json:"id"`
	Payload   []byte    `json:"payload"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}

func NewEvent(Type string, payload []byte) *Event {
	return &Event{
		ID:        uuid.New().String(),
		Payload:   payload,
		Type:      Type,
		Timestamp: time.Now(),
	}
}
