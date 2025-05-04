package repository

import (
	"github.com/Koyo-os/form-service/internal/entity"
	"github.com/Koyo-os/form-service/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Repository struct {
	db     *gorm.DB
	logger *logger.Logger
}

func Init(db *gorm.DB, logger *logger.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

func (repo *Repository) Create(form *entity.Form) error {
	res := repo.db.Create(form)

	if err := res.Error; err != nil {
		repo.logger.Error("error create form", zap.Error(err))

		return err
	}

	return nil
}

func (repo *Repository) UpdateDesc(formID uuid.UUID, desc string) error {
	res := repo.db.Where(&entity.Form{
		ID: formID,
	}).Update("description", desc)

	if err := res.Error; err != nil {
		repo.logger.Error("error update form", zap.String("form_id", formID.String()), zap.Error(err))

		return err
	}

	return nil
}

func (repo *Repository) UpdateAuthor(formID uuid.UUID, author string) error {
	res := repo.db.Where(&entity.Form{
		ID: formID,
	}).Update("author", author)

	if err := res.Error; err != nil {
		repo.logger.Error("error update form", zap.String("form_id", formID.String()), zap.Error(err))

		return err
	}

	return nil
}

func (repo *Repository) UpdateStatus(formID uuid.UUID, closed bool) error {
	res := repo.db.Where(&entity.Form{
		Closed: closed,
	}).Update("Closed", closed)

	if err := res.Error; err != nil {
		repo.logger.Error("error update form",
			zap.String("form_id", formID.String()),
			zap.Error(err),
		)

		return err
	}

	return nil
}

func (repo *Repository) UpdateQuestion(formID uuid.UUID, orderNumber uint, question *entity.Question) error {
	res := repo.db.Where(&entity.Question{
		FormID:      formID,
		OrderNumber: orderNumber,
	}).Updates(&entity.Question{
		Content:     question.Content,
		OrderNumber: question.OrderNumber,
	})

	if err := res.Error; err != nil {
		repo.logger.Error("error update question",
			zap.String("form_id", formID.String()),
			zap.Uint("order_number", orderNumber),
			zap.Error(err),
		)

		return err
	}

	return nil
}

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
