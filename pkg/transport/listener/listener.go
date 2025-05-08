package listener

import (
	"context"
	"encoding/json"

	"github.com/Koyo-os/Poll-service/internal/service"
	"github.com/Koyo-os/form-service/internal/entity"
	"github.com/Koyo-os/form-service/pkg/config"
	"github.com/Koyo-os/form-service/pkg/logger"
	"go.uber.org/zap"
)

type Listener struct {
	inputChan chan entity.Event
	logger    *logger.Logger
	service   service.PollService
	cfg       *config.Config
}

func Init(inputChan chan entity.Event, logger *logger.Logger, cfg *config.Config, service service.PollService) *Listener {
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
			if event.Type == list.cfg.Reqs.CreatePollRequestType {
				var poll entity.Poll

				if err := json.Unmarshal(event.Payload, &poll); err != nil {
					list.logger.Error("error unmarshal poll", zap.Error(err))
				}

				if err := list.service.Add(&poll); err != nil {
					list.logger.Error("error add poll to db", zap.Error(err))
				}
			} else if event.Type == list.cfg.Reqs.UpdatePollRequestType {
				var poll entity.Poll

				if err := json.Unmarshal(event.Payload, &poll); err != nil {
					list.logger.Error("error unmarshal poll", zap.Error(err))
				}

				if err := list.service.Update(poll.ID.String(), &poll); err != nil {
					list.logger.Error("error update poll", zap.Error(err))
				}
			} else {
				list.logger.Warn("unknown event type reciewed", zap.String("type", event.Type))
			}

		case <-ctx.Done():
			list.logger.Info("stopping listeners...")
			return
		}
	}
}
