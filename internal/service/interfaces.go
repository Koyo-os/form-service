package service

import (
	"context"

	"github.com/Koyo-os/form-service/internal/entity"
	"github.com/google/uuid"
)

type (
	Repository interface {
		Create(*entity.Form) error
		Update(uuid.UUID, string, any) error
		DeleteForm(uuid.UUID) error
		DeleteQuestion(uuid.UUID, uint) error
	}

	Publisher interface {
		Publish(any, string) error
	}

	Casher interface {
		DoCashing(ctx context.Context, key string, payload any) error // payload must to be pointer
		RemoveFromCash(ctx context.Context, key string) error
	}
)
