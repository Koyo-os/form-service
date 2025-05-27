package service

import (
	"context"
	"time"

	"github.com/Koyo-os/form-service/internal/entity"
	"github.com/Koyo-os/form-service/pkg/retrier"
	"github.com/google/uuid"
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
// It sets up a context with a default 10-second timeout for all service operations.
// Parameters:
//   - casher: Cache handler implementation
//   - repo: Repository implementation for data access
//   - publisher: Event publisher implementation
//
// Returns:
//   - *Service: Initialized service instance
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
// It performs the following operations:
//  1. Persists the form in the repository
//  2. Publishes a "form.created" event
//  3. Caches the form data (with retry logic)
//
// Parameters:
//   - form: Pointer to the Form entity to create
//
// Returns:
//   - error: Any error that occurs during the operation
func (s *Service) CreateForm(form *entity.Form) error {
	if err := s.repo.Create(form); err != nil {
		return err
	}

	cherr := make(chan error, 1)

	go func() {
		ctx, cancel := s.getContext()
		defer cancel()
		cherr <- retrier.Do(3, 5, func() error {
			return s.casher.AddToCash(ctx, form.ID.String(), form)
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

// CreateQuestion adds a new question to an existing form.
// It performs the following operations:
//  1. Persists the question in the repository
//  2. Retrieves the updated form
//  3. Publishes a "form.updated" event (with retry logic)
//  4. Updates the form in cache (with retry logic)
//
// Parameters:
//   - question: Pointer to the Question entity to create
//
// Returns:
//   - error: Any error that occurs during the operation
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
		ctx, cancel := s.getContext()
		defer cancel()

		return s.casher.AddToCash(ctx, form.ID.String(), form)
	}); err != nil {
		return err
	}

	return <-cherr
}

// UpdateStatus changes the closed/open status of a form.
// It performs the following operations:
//  1. Updates the status in the repository
//  2. Retrieves the updated form
//  3. Updates the form in cache (with retry logic)
//  4. Publishes a "form.created" event (note: potentially should be "form.updated")
//
// Parameters:
//   - form_id: String UUID of the form to update
//   - closed: Boolean indicating new status (true = closed)
//
// Returns:
//   - error: Any error that occurs during the operation
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
			ctx, cancel := s.getContext()
			defer cancel()

			return s.casher.AddToCash(ctx, form_id, form)
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

// Update modifies multiple fields of a form at once.
// It performs the following operations:
//  1. Updates fields in the repository
//  2. Updates the cache with new values (with retry logic)
//  3. Publishes a "form.updated" event (with retry logic)
//
// Parameters:
//   - formID: UUID of the form to update
//   - values: Interface containing the new field values
//
// Returns:
//   - error: Any error that occurs during the operation
func (s *Service) Update(formID uuid.UUID, values any) error {
	if err := s.repo.UpdateMany(formID, values); err != nil {
		return err
	}

	cherr := make(chan error, 1)

	go func() {
		ctx, cancel := s.getContext()
		defer cancel()

		cherr <- retrier.Do(3, 5, func() error {
			return s.casher.AddToCash(ctx, formID.String(), values)
		})
	}()

	if err := retrier.Do(3, 5, func() error {
		return s.publisher.Publish(values, "form.updated")
	}); err != nil {
		return err
	}

	return <-cherr
}

// UpdateDescription changes the description of a form.
// It performs the following operations:
//  1. Updates the description in the repository
//  2. Retrieves the updated form
//  3. Updates the form in cache (with retry logic)
//  4. Publishes a "form.updated" event (with retry logic)
//
// Parameters:
//   - formId: String UUID of the form to update
//   - desc: New description text
//
// Returns:
//   - error: Any error that occurs during the operation
func (s *Service) UpdateDescription(formId string, desc string) error {
	uid, err := uuid.Parse(formId)
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
		ctx, cancel := s.getContext()
		defer cancel()

		cherr <- retrier.Do(3, 5, func() error {
			return s.casher.AddToCash(ctx, formId, form)
		})
	}()

	if err = retrier.Do(3, 5, func() error {
		return s.publisher.Publish(form, "form.updated")
	}); err != nil {
		return err
	}

	return nil
}

// DeleteForm removes a form from the system.
// It performs the following operations:
//  1. Deletes the form from the repository
//  2. Removes the form from cache (with retry logic)
//  3. Publishes a "form.deleted" event (with retry logic)
//
// Parameters:
//   - formId: String UUID of the form to delete
//
// Returns:
//   - error: Any error that occurs during the operation
func (s *Service) DeleteForm(formId string) error {
	uid, err := uuid.Parse(formId)
	if err != nil {
		return err
	}

	if err = s.repo.DeleteForm(uid); err != nil {
		return err
	}

	cherr := make(chan error, 1)

	go func() {
		ctx, cancel := s.getContext()
		defer cancel()

		cherr <- retrier.Do(3, 5, func() error {
			return s.casher.RemoveFromCash(ctx, formId)
		})
	}()

	if err = retrier.Do(3, 5, func() error {
		return s.publisher.Publish(struct {
			FormID string
		}{
			FormID: formId,
		}, "form.deleted")
	}); err != nil {
		return err
	}

	return <-cherr
}

// DeleteQuestion removes a question from a form.
// It performs the following operations:
//  1. Deletes the question from the repository
//  2. Retrieves the updated form
//  3. Updates the form in cache (with retry logic)
//  4. Publishes a "form.updated" event
//
// Parameters:
//   - formId: String UUID of the form containing the question
//   - orderNumber: The order number of the question to delete
//
// Returns:
//   - error: Any error that occurs during the operation
func (s *Service) DeleteQuestion(formId string, orderNumber uint) error {
	uid, err := uuid.Parse(formId)
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
		ctx, cancel := s.getContext()
		defer cancel()

		cherr <- retrier.Do(3, 5, func() error {
			return s.casher.AddToCash(ctx, formId, form)
		})
	}()

	if err = s.publisher.Publish(form, "form.updated"); err != nil {
		return err
	}

	return <-cherr
}
