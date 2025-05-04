package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	Question struct {
		gorm.Model
		FormID      uuid.UUID
		Content     string
		OrderNumber uint
	}

	Form struct {
		gorm.Model
		ID          uuid.UUID
		Description string
		Closed      bool
		Questions   []Question
		Author      string
		CreatedAt   time.Time
	}

	OutputQuestion struct {
		Content     string `json:"content"`
		OrderNumber uint   `json:"order_number"`
	}

	OutputForm struct {
		ID          string           `json:"id"`
		Closed      bool             `json:"closed"`
		Description string           `json:"description"`
		Author      string           `json:"author"`
		CreatedAt   string           `json:"created_at"`
		Questions   []OutputQuestion `json:"questions"`
	}
)

func (o *Question) ToOutput() OutputQuestion {
	return OutputQuestion{
		Content:     o.Content,
		OrderNumber: o.OrderNumber,
	}
}

func (f *Form) ToOutput() OutputForm {
	return OutputForm{
		ID:          f.ID.String(),
		Description: f.Description,
		Author:      f.Author,
		CreatedAt:   f.CreatedAt.String(),
		Closed:      f.Closed,
	}
}

func (f *Form) ToJson() ([]byte, error) {
	form := f.ToOutput()
	form.Questions = make([]OutputQuestion, len(f.Questions))

	for i, fm := range f.Questions {
		form.Questions[i] = fm.ToOutput()
	}

	formJson, err := json.Marshal(&form)
	return formJson, err
}
