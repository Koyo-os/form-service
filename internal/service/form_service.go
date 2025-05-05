package service

import (
	"context"

	"github.com/Koyo-os/form-service/internal/entity"
	"github.com/Koyo-os/form-service/pkg/retrier"
	"github.com/google/uuid"
)

type Service struct {
	casher    Casher
	repo      Repository
	publisher Publisher

	ctx context.Context
}

func (s *Service) CreateForm(form *entity.Form) error {
	if err := s.repo.Create(form); err != nil {
		return err
	}

	cherr := make(chan error, 0)

	go func() {
		cherr <- retrier.Do(3, 5, func() error {
			return s.casher.DoCashing(s.ctx, form.ID.String(), form)
		})
	}()

	if err := s.publisher.Publish(form, "form.created"); err != nil {
		return err
	}

	if err := <-cherr; err != nil {
		return err
	}

	return nil
}

func (s *Service) UpdateStatus(form_id string, closed bool) error {
	uid, err := uuid.Parse(form_id)
	if err != nil {
		return err
	}

	if err = s.repo.Update(uid, "Closed", closed); err != nil {
		return err
	}

	cherr := make(chan error, 0)

	go func() {
		cherr <- retrier.Do(3, 5, func() error {
			return s.casher.DoCashing(s.ctx)
		})
	}()
}
