// Package entity defines the core data structures used throughout the application
package entity

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	// Question represents a single question within a form
	Question struct {
		gorm.Model
		FormID      uuid.UUID `gorm:"type:uuid"` // Reference to the parent form
		Content     string    // The actual question text
		OrderNumber uint      // Position of question in form
		Form        Form      `gorm:"foreignKey:FormID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"` // Relation to parent form
	}

	// Form represents a questionnaire or survey form
	Form struct {
		ID          uuid.UUID  `gorm:"type:uuid;primaryKey"` // Unique identifier
		Title       string     // Title of the form
		Description string     // Form description or purpose
		Closed      bool       // Whether form is closed for responses
		Questions   []Question `gorm:"foreignKey:FormID"` // Collection of form questions
		Author      string     // Creator of the form
		CreatedAt   time.Time  // Creation timestamp
	}

	// OutputQuestion is a DTO for question data in API responses
	OutputQuestion struct {
		Content     string `json:"content"`      // Question text
		OrderNumber uint   `json:"order_number"` // Question position
	}

	// OutputForm is a DTO for form data in API responses
	OutputForm struct {
		ID          string           `json:"id"`          // Form identifier
		Closed      bool             `json:"closed"`      // Form status
		Description string           `json:"description"` // Form description
		Author      string           `json:"author"`      // Form creator
		CreatedAt   string           `json:"created_at"`  // Creation time
		Questions   []OutputQuestion `json:"questions"`   // Form questions
	}
)

func (f *Form) Validate() error {
	if f.ID == uuid.Nil {
		return errors.New("form ID can not be nil")
	}
	if f.Author == "" {
		return errors.New("author ID can not be nil")
	}

	return nil
}

// ToOutput converts a Question entity to its DTO representation
func (o *Question) ToOutput() OutputQuestion {
	return OutputQuestion{
		Content:     o.Content,
		OrderNumber: o.OrderNumber,
	}
}

// ToOutput converts a Form entity to its DTO representation
func (f *Form) ToOutput() OutputForm {
	return OutputForm{
		ID:          f.ID.String(),
		Description: f.Description,
		Author:      f.Author,
		CreatedAt:   f.CreatedAt.String(),
		Closed:      f.Closed,
	}
}

// ToJson converts a Form entity to its JSON representation
// including all related questions
func (f *Form) ToJson() ([]byte, error) {
	form := f.ToOutput()
	form.Questions = make([]OutputQuestion, len(f.Questions))

	// Convert each question to its DTO form
	for i, fm := range f.Questions {
		form.Questions[i] = fm.ToOutput()
	}

	// Marshal the complete form to JSON
	formJson, err := json.Marshal(&form)
	return formJson, err
}
