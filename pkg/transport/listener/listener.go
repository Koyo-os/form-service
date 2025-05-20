package listener

import (
	"context"
	"encoding/json"

	"github.com/Koyo-os/form-service/internal/entity"
	"github.com/Koyo-os/form-service/internal/service"
	"github.com/Koyo-os/form-service/pkg/config"
	"github.com/Koyo-os/form-service/pkg/logger"
	"go.uber.org/zap"
)

type Listener struct {
	inputChan chan entity.Event
	logger    *logger.Logger
	service   *service.Service
	cfg       *config.Config
}

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

func (list *Listener) Listen(ctx context.Context) {
	for {
		select {
		case event := <-list.inputChan:
			switch event.Type {
			case list.cfg.Reqs.CreateRequestType:
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
			}

		case <-ctx.Done():
			list.logger.Info("stopping listeners...")
			return
		}
	}
}
