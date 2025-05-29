package entity

import (
	"errors"
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

func (e *Event) Validate() error {
	if e.ID == "" {
		return errors.New("event_id is nil")
	}

	if e.Payload == nil {
		return errors.New("payload is nil")
	}

	if e.Type == "" {
		return errors.New("type is nil")
	}

	return nil
}
