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

func (repo *Repository) Create(payload any) error {
	res := repo.db.Create(payload)

	if err := res.Error; err != nil {
		repo.logger.Error("error create entity", zap.Error(err))

		return err
	}

	return nil
}

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
