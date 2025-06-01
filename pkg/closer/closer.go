package closer

import (
	"github.com/Koyo-os/form-service/pkg/logger"
	"go.uber.org/zap"
)

type (
	Closer interface {
		Close() error
	}

	CloserGroup struct {
		closers []Closer
		logger  *logger.Logger
	}
)

func NewCloserGroup(logger *logger.Logger, closers ...Closer) *CloserGroup {
	return &CloserGroup{
		closers: closers,
		logger:  logger,
	}
}

func (c *CloserGroup) Close() error {
	var err error

	for _, closer := range c.closers {
		if err := closer.Close(); err != nil {
			c.logger.Error("failed close", zap.Error(err))
		}
	}
	return err
}
