// Package repository provides data persistence functionality using GORM
package repository

import (
	"github.com/Koyo-os/form-service/internal/entity"
	"github.com/Koyo-os/form-service/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Repository handles database operations using GORM
type Repository struct {
	db     *gorm.DB
	logger *logger.Logger
}

// Init creates and returns a new Repository instance
func Init(db *gorm.DB, logger *logger.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

// Create persists a new entity in the database
// Parameters:
//   - payload: Any struct that maps to a database table
//
// Returns error if the creation fails
func (repo *Repository) Create(payload any) error {
	res := repo.db.Create(payload)

	if err := res.Error; err != nil {
		repo.logger.Error("error create entity", zap.Error(err))
		return err
	}

	return nil
}

// Get retrieves a form by its ID
// Parameters:
//   - ID: UUID of the form to retrieve
//
// Returns:
//   - *entity.Form: Retrieved form or nil if not found
//   - error: Any error that occurred during retrieval
func (repo *Repository) Get(ID uuid.UUID) (*entity.Form, error) {
	var form entity.Form

	res := repo.db.Where("ID = ?", ID).First(&form)
	if err := res.Error; err != nil {
		repo.logger.Error("error get form",
			zap.String("form_id", ID.String()),
			zap.Error(err),
		)
		return nil, err
	}

	return &form, nil
}

// Update modifies a single column of a form
// Parameters:
//   - ID: UUID of the form to update
//   - key: Column name to update
//   - value: New value for the column
//
// Returns error if the update fails
func (repo *Repository) Update(ID uuid.UUID, key string, value any) error {
	res := repo.db.Where("ID = ?", ID).Update(key, value)

	if err := res.Error; err != nil {
		repo.logger.Error("error update form",
			zap.String("form_id", ID.String()),
			zap.Error(err),
		)
		return err
	}

	return nil
}

// UpdateMany updates multiple columns of a form simultaneously
// Parameters:
//   - ID: UUID of the form to update
//   - value: Struct containing the columns and values to update
//
// Returns error if the update fails
func (repo *Repository) UpdateMany(ID uuid.UUID, value any) error {
	res := repo.db.Where("ID = ?", ID).Updates(value)

	if err := res.Error; err != nil {
		repo.logger.Error("error update many",
			zap.String("id", ID.String()),
			zap.Error(err))
		return err
	}

	return nil
}

// UpdateQuestion modifies a single column of a question
// Parameters:
//   - id: UUID of the question to update
//   - key: Column name to update
//   - value: New value for the column
//
// Returns error if the update fails
func (repo *Repository) UpdateQuestion(id uuid.UUID, key string, value any) error {
	res := repo.db.Where("ID = ?", id).Update(key, value)

	if err := res.Error; err != nil {
		repo.logger.Error("error update question",
			zap.String("column", key),
			zap.String("question_id", id.String()),
			zap.Error(err))
		return err
	}

	return nil
}

// UpdateQuestionMany updates multiple columns of a question simultaneously
// Parameters:
//   - id: UUID of the question to update
//   - value: Struct containing the columns and values to update
//
// Returns error if the update fails
func (repo *Repository) UpdateQuestionMany(id uuid.UUID, value any) error {
	res := repo.db.Where("ID = ?", id).Updates(value)

	if err := res.Error; err != nil {
		repo.logger.Error("error update question many",
			zap.String("question_id", id.String()),
			zap.Error(err))
		return err
	}

	return nil
}

// DeleteForm removes a form from the database
// Parameters:
//   - formID: UUID of the form to delete
//
// Returns error if the deletion fails
func (repo *Repository) DeleteForm(formID uuid.UUID) error {
	res := repo.db.Where(&entity.Form{
		ID: formID,
	}).Delete(&entity.Form{})

	if err := res.Error; err != nil {
		repo.logger.Error("error delete form",
			zap.String("form_id", formID.String()),
			zap.Error(err),
		)
		return err
	}

	return nil
}

// DeleteQuestion removes a question from a form
// Parameters:
//   - formID: UUID of the form containing the question
//   - orderNumber: Position of the question in the form
//
// Returns error if the deletion fails
func (repo *Repository) DeleteQuestion(formID uuid.UUID, orderNumber uint) error {
	res := repo.db.Where(&entity.Question{
		FormID:      formID,
		OrderNumber: orderNumber,
	}).Delete(&entity.Question{})

	if err := res.Error; err != nil {
		repo.logger.Error("error delete question",
			zap.String("form_id", formID.String()),
			zap.Uint("order_number", orderNumber),
			zap.Error(err),
		)
		return err
	}

	return nil
}
