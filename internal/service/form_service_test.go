package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Koyo-os/form-service/internal/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCasher is a mock implementation of the Casher interface
type MockCasher struct {
	mock.Mock
}

func (m *MockCasher) AddToCash(ctx context.Context, key string, value interface{}) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

func (m *MockCasher) RemoveFromCash(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCasher) GetCashFor(ctx context.Context, key string) ([]byte, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

// MockRepository is a mock implementation of the Repository interface
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(entity interface{}) error {
	args := m.Called(entity)
	return args.Error(0)
}

func (m *MockRepository) Get(id uuid.UUID) (*entity.Form, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Form), args.Error(1)
}

func (m *MockRepository) Update(id uuid.UUID, field string, value interface{}) error {
	args := m.Called(id, field, value)
	return args.Error(0)
}

func (m *MockRepository) UpdateMany(id uuid.UUID, values interface{}) error {
	args := m.Called(id, values)
	return args.Error(0)
}

func (m *MockRepository) DeleteForm(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRepository) DeleteQuestion(formID uuid.UUID, orderNumber uint) error {
	args := m.Called(formID, orderNumber)
	return args.Error(0)
}

// MockPublisher is a mock implementation of the Publisher interface
type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) Publish(data interface{}, event string) error {
	args := m.Called(data, event)
	return args.Error(0)
}

func setupService() (*Service, *MockCasher, *MockRepository, *MockPublisher) {
	mockCasher := &MockCasher{}
	mockRepo := &MockRepository{}
	mockPublisher := &MockPublisher{}
	service := Init(mockCasher, mockRepo, mockPublisher, 5*time.Second)
	return service, mockCasher, mockRepo, mockPublisher
}

func TestService_CreateForm_Success(t *testing.T) {
	service, mockCasher, mockRepo, mockPublisher := setupService()

	form := &entity.Form{
		ID:          uuid.New(),
		Title:       "Test Form",
		Description: "Test Description",
	}

	mockRepo.On("Create", form).Return(nil)
	mockCasher.On("AddToCash", mock.AnythingOfType("*context.timerCtx"), form.ID.String(), form).
		Return(nil)
	mockPublisher.On("Publish", form, "form.created").Return(nil)

	err := service.CreateForm(form)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockCasher.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestService_CreateForm_NilForm(t *testing.T) {
	service, _, _, _ := setupService()

	err := service.CreateForm(nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "form cannot be nil")
}

func TestService_CreateForm_RepositoryError(t *testing.T) {
	service, _, mockRepo, _ := setupService()

	form := &entity.Form{
		ID:    uuid.New(),
		Title: "Test Form",
	}

	mockRepo.On("Create", form).Return(errors.New("database error"))

	err := service.CreateForm(form)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create form in repository")
	mockRepo.AssertExpectations(t)
}

func TestService_CreateForm_CacheError(t *testing.T) {
	service, mockCasher, mockRepo, mockPublisher := setupService()

	form := &entity.Form{
		ID:    uuid.New(),
		Title: "Test Form",
	}

	mockRepo.On("Create", form).Return(nil)
	mockCasher.On("AddToCash", mock.AnythingOfType("*context.timerCtx"), form.ID.String(), form).
		Return(errors.New("cache error"))
	mockPublisher.On("Publish", form, "form.created").Return(nil)

	err := service.CreateForm(form)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cache error")
}

func TestService_CreateQuestion_Success(t *testing.T) {
	service, mockCasher, mockRepo, mockPublisher := setupService()

	formID := uuid.New()
	question := &entity.Question{
		FormID:      formID,
		OrderNumber: 1,
	}

	form := &entity.Form{
		ID:    formID,
		Title: "Test Form",
	}

	mockRepo.On("Create", question).Return(nil)
	mockRepo.On("Get", formID).Return(form, nil)
	mockCasher.On("AddToCash", mock.AnythingOfType("*context.timerCtx"), formID.String(), form).
		Return(nil)
	mockPublisher.On("Publish", form, "form.updated").Return(nil)

	err := service.CreateQuestion(question)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockCasher.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestService_CreateQuestion_NilQuestion(t *testing.T) {
	service, _, _, _ := setupService()

	err := service.CreateQuestion(nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "question cannot be nil")
}

func TestService_CreateQuestion_RepositoryError(t *testing.T) {
	service, _, mockRepo, _ := setupService()

	question := &entity.Question{
		FormID: uuid.New(),
	}

	mockRepo.On("Create", question).Return(errors.New("database error"))

	err := service.CreateQuestion(question)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create question in repository")
}

func TestService_CreateQuestion_FormRetrievalError(t *testing.T) {
	service, _, mockRepo, _ := setupService()

	formID := uuid.New()
	question := &entity.Question{
		FormID: formID,
	}

	mockRepo.On("Create", question).Return(nil)
	mockRepo.On("Get", formID).Return(nil, errors.New("form not found"))

	err := service.CreateQuestion(question)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to retrieve updated form")
}

func TestService_UpdateStatus_Success(t *testing.T) {
	service, mockCasher, mockRepo, mockPublisher := setupService()

	formID := uuid.New()
	form := &entity.Form{
		ID:     formID,
		Title:  "Test Form",
		Closed: true,
	}

	mockRepo.On("Update", formID, "Closed", true).Return(nil)
	mockRepo.On("Get", formID).Return(form, nil)
	mockCasher.On("AddToCash", mock.AnythingOfType("*context.timerCtx"), formID.String(), form).
		Return(nil)
	mockPublisher.On("Publish", form, "form.updated").Return(nil)

	err := service.UpdateStatus(formID, true)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockCasher.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestService_UpdateStatus_RepositoryError(t *testing.T) {
	service, _, mockRepo, _ := setupService()

	formID := uuid.New()

	mockRepo.On("Update", formID, "Closed", false).Return(errors.New("database error"))

	err := service.UpdateStatus(formID, false)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update form status in repository")
}

func TestService_Update_Success(t *testing.T) {
	service, mockCasher, mockRepo, mockPublisher := setupService()

	formID := uuid.New()
	values := map[string]interface{}{
		"Title": "Updated Title",
	}
	form := &entity.Form{
		ID:    formID,
		Title: "Updated Title",
	}

	mockRepo.On("UpdateMany", formID, values).Return(nil)
	mockRepo.On("Get", formID).Return(form, nil)
	mockCasher.On("AddToCash", mock.AnythingOfType("*context.timerCtx"), formID.String(), form).
		Return(nil)
	mockPublisher.On("Publish", form, "form.updated").Return(nil)

	err := service.Update(formID, values)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockCasher.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestService_Update_NilValues(t *testing.T) {
	service, _, _, _ := setupService()

	formID := uuid.New()

	err := service.Update(formID, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "values cannot be nil")
}

func TestService_Update_RepositoryError(t *testing.T) {
	service, _, mockRepo, _ := setupService()

	formID := uuid.New()
	values := map[string]interface{}{
		"Title": "Updated Title",
	}

	mockRepo.On("UpdateMany", formID, values).Return(errors.New("database error"))

	err := service.Update(formID, values)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update form in repository")
}

func TestService_UpdateDescription_Success(t *testing.T) {
	service, mockCasher, mockRepo, mockPublisher := setupService()

	formID := uuid.New()
	description := "Updated Description"
	form := &entity.Form{
		ID:          formID,
		Title:       "Test Form",
		Description: description,
	}

	mockRepo.On("Update", formID, "Description", description).Return(nil)
	mockRepo.On("Get", formID).Return(form, nil)
	mockCasher.On("AddToCash", mock.AnythingOfType("*context.timerCtx"), formID.String(), form).
		Return(nil)
	mockPublisher.On("Publish", form, "form.updated").Return(nil)

	err := service.UpdateDescription(formID, description)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockCasher.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestService_UpdateDescription_RepositoryError(t *testing.T) {
	service, _, mockRepo, _ := setupService()

	formID := uuid.New()
	description := "Updated Description"

	mockRepo.On("Update", formID, "Description", description).Return(errors.New("database error"))

	err := service.UpdateDescription(formID, description)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update form description in repository")
}

func TestService_DeleteForm_Success(t *testing.T) {
	service, mockCasher, mockRepo, mockPublisher := setupService()

	formID := uuid.New()

	mockRepo.On("DeleteForm", formID).Return(nil)
	mockCasher.On("RemoveFromCash", mock.AnythingOfType("*context.timerCtx"), formID.String()).
		Return(nil)
	mockPublisher.On("Publish", mock.MatchedBy(func(data interface{}) bool {
		if payload, ok := data.(struct {
			FormID string `json:"form_id"`
		}); ok {
			return payload.FormID == formID.String()
		}
		return false
	}), "form.deleted").Return(nil)

	err := service.DeleteForm(formID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockCasher.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestService_DeleteForm_RepositoryError(t *testing.T) {
	service, _, mockRepo, _ := setupService()

	formID := uuid.New()

	mockRepo.On("DeleteForm", formID).Return(errors.New("database error"))

	err := service.DeleteForm(formID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete form from repository")
}
