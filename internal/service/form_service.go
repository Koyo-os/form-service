package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Koyo-os/form-service/internal/entity"
	"github.com/Koyo-os/form-service/pkg/retrier"
	"github.com/google/uuid"
)

const (
	DefaultRetryAttempts = 3
	DefaultRetryDelay    = 5
)

// Service provides business logic for form management operations.
// It coordinates between repository, cache, and event publishing systems.
type Service struct {
	casher    Casher     // Handles caching operations for forms
	repo      Repository // Provides persistence layer access
	publisher Publisher  // Manages event publishing
	timeout   time.Duration
}

// Init initializes and returns a new Service instance with dependencies.
func Init(casher Casher, repo Repository, publisher Publisher, timeout time.Duration) *Service {
	return &Service{
		casher:    casher,
		repo:      repo,
		publisher: publisher,
		timeout:   timeout,
	}
}

func (s *Service) getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), s.timeout)
}

// CreateForm creates a new form in the system.
func (s *Service) CreateForm(form *entity.Form) error {
	if form == nil {
		return errors.New("form cannot be nil")
	}

	// 1. Critical operation first (database)
	if err := s.repo.Create(form); err != nil {
		return fmt.Errorf("failed to create form in repository: %w", err)
	}

	// 2. Run non-critical operations concurrently
	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	// Cache operation
	wg.Add(1)
	go func() {
		defer wg.Done()
		ctx, cancel := s.getContext()
		defer cancel()

		if err := retrier.Do(DefaultRetryAttempts, DefaultRetryDelay, func() error {
			return s.casher.AddToCash(ctx, form.ID.String(), form)
		}); err != nil {
			errChan <- fmt.Errorf("cache error: %w", err)
		}
	}()

	// Publish operation
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := retrier.Do(DefaultRetryAttempts, DefaultRetryDelay, func() error {
			return s.publisher.Publish(form, "form.created")
		}); err != nil {
			errChan <- fmt.Errorf("publish error: %w", err)
		}
	}()

	wg.Wait()
	close(errChan)

	// Return first error if any
	for err := range errChan {
		return err
	}

	return nil
}

// CreateQuestion adds a new question to an existing form.
func (s *Service) CreateQuestion(question *entity.Question) error {
	if question == nil {
		return errors.New("question cannot be nil")
	}

	// 1. Critical operation first (database)
	if err := s.repo.Create(question); err != nil {
		return fmt.Errorf("failed to create question in repository: %w", err)
	}

	// 2. Get updated form
	form, err := s.repo.Get(question.FormID)
	if err != nil {
		return fmt.Errorf("failed to retrieve updated form: %w", err)
	}

	// 3. Run non-critical operations concurrently
	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	// Cache operation
	wg.Add(1)
	go func() {
		defer wg.Done()
		ctx, cancel := s.getContext()
		defer cancel()

		if err := retrier.Do(DefaultRetryAttempts, DefaultRetryDelay, func() error {
			return s.casher.AddToCash(ctx, form.ID.String(), form)
		}); err != nil {
			errChan <- fmt.Errorf("cache error: %w", err)
		}
	}()

	// Publish operation
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := retrier.Do(DefaultRetryAttempts, DefaultRetryDelay, func() error {
			return s.publisher.Publish(form, "form.updated")
		}); err != nil {
			errChan <- fmt.Errorf("publish error: %w", err)
		}
	}()

	wg.Wait()
	close(errChan)

	// Return first error if any
	for err := range errChan {
		return err
	}

	return nil
}

// UpdateStatus changes the closed/open status of a form.
func (s *Service) UpdateStatus(formID uuid.UUID, closed bool) error {
	// 1. Critical operation first (database)
	if err := s.repo.Update(formID, "Closed", closed); err != nil {
		return fmt.Errorf("failed to update form status in repository: %w", err)
	}

	// 2. Get updated form
	form, err := s.repo.Get(formID)
	if err != nil {
		return fmt.Errorf("failed to retrieve updated form: %w", err)
	}

	// 3. Run non-critical operations concurrently
	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	// Cache operation
	wg.Add(1)
	go func() {
		defer wg.Done()
		ctx, cancel := s.getContext()
		defer cancel()

		if err := retrier.Do(DefaultRetryAttempts, DefaultRetryDelay, func() error {
			return s.casher.AddToCash(ctx, formID.String(), form)
		}); err != nil {
			errChan <- fmt.Errorf("cache error: %w", err)
		}
	}()

	// Publish operation
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := retrier.Do(DefaultRetryAttempts, DefaultRetryDelay, func() error {
			return s.publisher.Publish(form, "form.updated")
		}); err != nil {
			errChan <- fmt.Errorf("publish error: %w", err)
		}
	}()

	wg.Wait()
	close(errChan)

	// Return first error if any
	for err := range errChan {
		return err
	}

	return nil
}

// Update modifies multiple fields of a form at once.
func (s *Service) Update(formID uuid.UUID, values any) error {
	if values == nil {
		return errors.New("values cannot be nil")
	}

	// 1. Critical operation first (database)
	if err := s.repo.UpdateMany(formID, values); err != nil {
		return fmt.Errorf("failed to update form in repository: %w", err)
	}

	// 2. Get updated form to ensure cache consistency
	form, err := s.repo.Get(formID)
	if err != nil {
		return fmt.Errorf("failed to retrieve updated form: %w", err)
	}

	// 3. Run non-critical operations concurrently
	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	// Cache operation (cache the complete form, not just values)
	wg.Add(1)
	go func() {
		defer wg.Done()
		ctx, cancel := s.getContext()
		defer cancel()

		if err := retrier.Do(DefaultRetryAttempts, DefaultRetryDelay, func() error {
			return s.casher.AddToCash(ctx, formID.String(), form)
		}); err != nil {
			errChan <- fmt.Errorf("cache error: %w", err)
		}
	}()

	// Publish operation (publish the complete form, not just values)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := retrier.Do(DefaultRetryAttempts, DefaultRetryDelay, func() error {
			return s.publisher.Publish(form, "form.updated")
		}); err != nil {
			errChan <- fmt.Errorf("publish error: %w", err)
		}
	}()

	wg.Wait()
	close(errChan)

	// Return first error if any
	for err := range errChan {
		return err
	}

	return nil
}

// UpdateDescription changes the description of a form.
func (s *Service) UpdateDescription(formID uuid.UUID, desc string) error {
	// 1. Critical operation first (database)
	if err := s.repo.Update(formID, "Description", desc); err != nil {
		return fmt.Errorf("failed to update form description in repository: %w", err)
	}

	// 2. Get updated form
	form, err := s.repo.Get(formID)
	if err != nil {
		return fmt.Errorf("failed to retrieve updated form: %w", err)
	}

	// 3. Run non-critical operations concurrently
	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	// Cache operation
	wg.Add(1)
	go func() {
		defer wg.Done()
		ctx, cancel := s.getContext()
		defer cancel()

		if err := retrier.Do(DefaultRetryAttempts, DefaultRetryDelay, func() error {
			return s.casher.AddToCash(ctx, formID.String(), form)
		}); err != nil {
			errChan <- fmt.Errorf("cache error: %w", err)
		}
	}()

	// Publish operation
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := retrier.Do(DefaultRetryAttempts, DefaultRetryDelay, func() error {
			return s.publisher.Publish(form, "form.updated")
		}); err != nil {
			errChan <- fmt.Errorf("publish error: %w", err)
		}
	}()

	wg.Wait()
	close(errChan)

	// Return first error if any
	for err := range errChan {
		return err
	}

	return nil
}

// DeleteForm removes a form from the system.
func (s *Service) DeleteForm(formID uuid.UUID) error {
	// 1. Critical operation first (database)
	if err := s.repo.DeleteForm(formID); err != nil {
		return fmt.Errorf("failed to delete form from repository: %w", err)
	}

	// 2. Run non-critical operations concurrently
	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	// Cache removal operation
	wg.Add(1)
	go func() {
		defer wg.Done()
		ctx, cancel := s.getContext()
		defer cancel()

		if err := retrier.Do(DefaultRetryAttempts, DefaultRetryDelay, func() error {
			return s.casher.RemoveFromCash(ctx, formID.String())
		}); err != nil {
			errChan <- fmt.Errorf("cache removal error: %w", err)
		}
	}()

	// Publish operation
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := retrier.Do(DefaultRetryAttempts, DefaultRetryDelay, func() error {
			return s.publisher.Publish(struct {
				FormID string `json:"form_id"`
			}{
				FormID: formID.String(),
			}, "form.deleted")
		}); err != nil {
			errChan <- fmt.Errorf("publish error: %w", err)
		}
	}()

	wg.Wait()
	close(errChan)

	// Return first error if any
	for err := range errChan {
		return err
	}

	return nil
}

// DeleteQuestion removes a question from a form.
func (s *Service) DeleteQuestion(formID uuid.UUID, orderNumber uint) error {
	// 1. Critical operation first (database)
	if err := s.repo.DeleteQuestion(formID, orderNumber); err != nil {
		return fmt.Errorf("failed to delete question from repository: %w", err)
	}

	// 2. Get updated form
	form, err := s.repo.Get(formID)
	if err != nil {
		return fmt.Errorf("failed to retrieve updated form: %w", err)
	}

	// 3. Run non-critical operations concurrently
	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	// Cache operation
	wg.Add(1)
	go func() {
		defer wg.Done()
		ctx, cancel := s.getContext()
		defer cancel()

		if err := retrier.Do(DefaultRetryAttempts, DefaultRetryDelay, func() error {
			return s.casher.AddToCash(ctx, formID.String(), form)
		}); err != nil {
			errChan <- fmt.Errorf("cache error: %w", err)
		}
	}()

	// Publish operation
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := retrier.Do(DefaultRetryAttempts, DefaultRetryDelay, func() error {
			return s.publisher.Publish(form, "form.updated")
		}); err != nil {
			errChan <- fmt.Errorf("publish error: %w", err)
		}
	}()

	wg.Wait()
	close(errChan)

	// Return first error if any
	for err := range errChan {
		return err
	}

	return nil
}
