package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Koyo-os/form-service/internal/entity"
	"github.com/Koyo-os/form-service/internal/repository"
	"github.com/Koyo-os/form-service/internal/service"
	"github.com/Koyo-os/form-service/pkg/closer"
	"github.com/Koyo-os/form-service/pkg/config"
	"github.com/Koyo-os/form-service/pkg/health"
	"github.com/Koyo-os/form-service/pkg/logger"
	"github.com/Koyo-os/form-service/pkg/retrier"
	"github.com/Koyo-os/form-service/pkg/transport/casher"
	"github.com/Koyo-os/form-service/pkg/transport/consumer"
	"github.com/Koyo-os/form-service/pkg/transport/listener"
	"github.com/Koyo-os/form-service/pkg/transport/publisher"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	var eventChan chan entity.Event

	logCfg := logger.Config{
		LogFile:   "app.log",
		LogLevel:  "debug",
		AppName:   "form-service",
		AddCaller: true,
	}

	if err := logger.Init(logCfg); err != nil {
		panic(err)
	}

	defer logger.Sync()

	logger := logger.Get()

	cfg, err := config.Init("config.yaml")
	if err != nil {
		logger.Error("error init config",
			zap.String("path", "config.yaml"),
			zap.Error(err))
		return
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	logger.Info("connecting to mariadb...", zap.String("dsn", dsn))

	db, err := retrier.Connect(10, 10, func() (*gorm.DB, error) {
		return gorm.Open(mysql.Open(dsn))
	})
	if err != nil {
		logger.Error("error initialyze database",
			zap.String("dsn", dsn),
			zap.Error(err))
	}

	logger.Info("connected to mariadb", zap.String("dsn", dsn))

	repo := repository.Init(db, logger)

	rabbitmqConns, err := retrier.MultiConnects(2, func() (*amqp.Connection, error) {
		return amqp.Dial(cfg.Urls.Rabbitmq)
	}, &retrier.RetrierOpts{Count: 3, Interval: 5})
	if err != nil {
		logger.Error("error connect to rabbitmq",
			zap.String("url", cfg.Urls.Rabbitmq),
			zap.Error(err))

		return
	}

	publish, err := publisher.Init(cfg, logger, rabbitmqConns[0])
	if err != nil {
		logger.Error("error initialize publisher", zap.Error(err))

		return
	}

	cons, err := consumer.Init(cfg, logger, rabbitmqConns[1])
	if err != nil {
		logger.Error("error initialize consumer", zap.Error(err))

		return
	}

	redisConn, err := retrier.Connect(3, 5, func() (*redis.Client, error) {
		client := redis.NewClient(&redis.Options{
			Addr:     cfg.Urls.Redis,
			DB:       0,
			Password: "",
		})

		return client, client.Ping(context.Background()).Err()
	})
	if err != nil {
		logger.Error("error connect to redis", zap.Error(err))

		return
	}

	cashers := casher.Init(redisConn, logger)

	core := service.Init(cashers, repo, publish, 10*time.Second)

	list := listener.Init(eventChan, logger, cfg, core)

	if err = cons.Subscribe(cfg.Exchange.Request, "request.*", cfg.Queue.Request); err != nil {
		logger.Error("error subscribe to queue", zap.Error(err))
		return
	}

	closers := closer.NewCloserGroup(cashers, list, cons, publish)
	health := health.NewHealthChecker(publish, cashers, cons)

	go health.StartHealthCheckServer("8080")
	go list.Listen(context.Background())
	go cons.ConsumeMessages(eventChan)

	<-signalChan
	logger.Info("Shutting down...")

	if err = closers.Close();err != nil{
		logger.Error("error closed", zap.Error(err))

		return
	}
}
