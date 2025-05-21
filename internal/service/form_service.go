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

	cherr := make(chan error, 1)

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

func (s *Service) CreateQuestion(question *entity.Question) error {
	if err := s.repo.Create(question); err != nil {
		return err
	}

	form, err := s.repo.Get(question.FormID)
	if err != nil {
		return err
	}

	cherr := make(chan error, 1)

	go func() {
		cherr <- retrier.Do(3, 5, func() error {
			return s.publisher.Publish(form, "form.updated")
		})
	}()

	if err = retrier.Do(3, 5, func() error {
		return s.casher.DoCashing(s.ctx, form.ID.String(), form)
	}); err != nil {
		return err
	}

	return <-cherr
}

func (s *Service) UpdateStatus(form_id string, closed bool) error {
	uid, err := uuid.Parse(form_id)
	if err != nil {
		return err
	}

	if err = s.repo.Update(uid, "Closed", closed); err != nil {
		return err
	}

	form, err := s.repo.Get(uid)
	if err != nil {
		return err
	}

	cherr := make(chan error, 1)

	go func() {
		cherr <- retrier.Do(3, 5, func() error {
			return s.casher.DoCashing(s.ctx, form_id, form)
		})
	}()

	if err = retrier.Do(3, 5, func() error {
		return s.publisher.Publish(form, "form.created")
	}); err != nil {
		return err
	}

	if err = <-cherr; err != nil {
		return err
	}

	return nil
}

func (s Service) Update(form_id uuid.UUID, values any) error {
	if err := s.repo.UpdateMany(form_id, values); err != nil {
		return err
	}

	cherr := make(chan error, 1)

	go func() {
		cherr <- retrier.Do(3, 5, func() error {
			return s.casher.DoCashing(s.ctx, form_id.String(), values)
		})
	}()

	if err := retrier.Do(3, 5, func() error {
		return s.publisher.Publish(values, "form.updated")
	}); err != nil {
		return err
	}

	return <-cherr
}

func (s *Service) UpdateDescription(form_id string, desc string) error {
	uid, err := uuid.Parse(form_id)
	if err != nil {
		return err
	}

	if err = s.repo.Update(uid, "Description", desc); err != nil {
		return err
	}

	form, err := s.repo.Get(uid)
	if err != nil {
		return err
	}

	cherr := make(chan error, 1)

	go func() {
		cherr <- retrier.Do(3, 5, func() error {
			return s.casher.DoCashing(s.ctx, form_id, form)
		})
	}()

	if err = retrier.Do(3, 5, func() error {
		return s.publisher.Publish(form, "form.updated")
	}); err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteForm(form_id string) error {
	uid, err := uuid.Parse(form_id)
	if err != nil {
		return err
	}

	if err = s.repo.DeleteForm(uid); err != nil {
		return err
	}

	cherr := make(chan error, 1)

	go func() {
		cherr <- retrier.Do(3, 5, func() error {
			return s.casher.RemoveFromCash(s.ctx, form_id)
		})
	}()

	if err = retrier.Do(3, 5, func() error {
		return s.publisher.Publish(struct {
			FormID string
		}{
			FormID: form_id,
		}, "form.deleted")
	}); err != nil {
		return err
	}

	return <-cherr
}

func (s *Service) DeleteQuestion(form_id string, orderNumber uint) error {
	uid, err := uuid.Parse(form_id)
	if err != nil {
		return err
	}

	if err = s.repo.DeleteQuestion(uid, orderNumber); err != nil {
		return err
	}

	form, err := s.repo.Get(uid)
	if err != nil {
		return err
	}

	cherr := make(chan error, 1)
	go func() {
		cherr <- retrier.Do(3, 5, func() error {
			return s.casher.DoCashing(s.ctx, form_id, form)
		})
	}()

	if err = s.publisher.Publish(form, "form.updated"); err != nil {
		return err
	}

	return <-cherr
}
