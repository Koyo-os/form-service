// Package listener provides event handling functionality for form-related operations
package listener

import (
	"context"
	"encoding/json"

	"github.com/Koyo-os/form-service/internal/entity"
	"github.com/Koyo-os/form-service/internal/service"
	"github.com/Koyo-os/form-service/pkg/config"
	"github.com/Koyo-os/form-service/pkg/logger"
	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Listener handles incoming events and routes them to appropriate service methods
type Listener struct {
	inputChan chan entity.Event // Channel for receiving events
	logger    *logger.Logger    // Logger for error tracking
	service   *service.Service  // Service layer for business logic
	cfg       *config.Config    // Application configuration
}

// Init creates a new Listener instance with all required dependencies
func Init(
	inputChan chan entity.Event,
	logger *logger.Logger,
	cfg *config.Config,
	service *service.Service,
) *Listener {
	return &Listener{
		inputChan: inputChan,
		service:   service,
		logger:    logger,
		cfg:       cfg,
	}
}

func (list *Listener) Close() error {
	close(list.inputChan)

	return nil
}

// Listen starts the event listening loop
// It processes incoming events based on their type and routes them to appropriate handlers
// The loop continues until the context is cancelled
func (list *Listener) Listen(ctx context.Context) {
	for {
		select {
		case event := <-list.inputChan:
			switch event.Type {
			case list.cfg.Reqs.CreateRequestType:
				// Handle form creation events
				form := new(entity.Form)

				if err := json.Unmarshal(event.Payload, &form); err != nil {
					list.logger.Error("error unmarshal event payload to form",
						zap.String("event_type", event.Type),
						zap.String("event_id", event.ID),
						zap.Error(err))
					continue
				}

				if err := list.service.CreateForm(form); err != nil {
					list.logger.Error("error create form", zap.Error(err))
					continue
				}

			case list.cfg.Reqs.UpdateRequestType:
				// Handle form update events
				form := new(entity.Form)

				if err := json.Unmarshal(event.Payload, &form); err != nil {
					list.logger.Error("error unmarshal payload to form",
						zap.String("event_id", event.ID),
						zap.String("event_type", event.Type),
						zap.Error(err))
					continue
				}

				if err := list.service.Update(form.ID, form); err != nil {
					list.logger.Error("error update form",
						zap.String("event_id", event.ID),
						zap.String("form_id", form.ID.String()),
						zap.Error(err))
					continue
				}

			case list.cfg.Reqs.DeleteFormRequestType:
				// Handle form deletion events
				req := new(struct {
					FormID string `json:"form_id"`
				})

				if err := sonic.Unmarshal(event.Payload, req); err != nil {
					list.logger.Error("error unmarshal request from event payload",
						zap.String("event_id", event.ID),
						zap.String("event_type", event.Type),
						zap.Error(err))
					continue
				}

				id, err := uuid.Parse(req.FormID)
				if err != nil {
					list.logger.Error("error parse form id",
						zap.String("event_id", event.ID),
						zap.String("event_type", event.Type),
						zap.Error(err))
					continue
				}

				if err = list.service.DeleteForm(id); err != nil {
					list.logger.Error("error delete form",
						zap.String("event_id", event.ID),
						zap.String("form_id", req.FormID),
						zap.Error(err))
					continue
				}
			}

		case <-ctx.Done():
			list.logger.Info("stopping listeners...")
			return
		}
	}
}
