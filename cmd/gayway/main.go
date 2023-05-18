package main

import (
	"fmt"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/alexdrl/zerowater"
	"github.com/kachvame/gayway/gateway"
	"github.com/kachvame/gayway/kv/etcd"
	gaywayLog "github.com/kachvame/gayway/log"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
)

func run() error {
	discordToken := os.Getenv("GAYWAY_DISCORD_TOKEN")
	etcdAddress := os.Getenv("GAYWAY_ETCD_ADDRESS")
	etcdPassword := os.Getenv("GAYWAY_ETCD_PASSWORD")
	etcdUsername := os.Getenv("GAYWAY_ETCD_USERNAME")
	kafkaAddress := os.Getenv("GAYWAY_KAFKA_ADDRESS")
	logLevel := os.Getenv("GAYWAY_LOG_LEVEL")
	dev := os.Getenv("GAYWAY_DEV")

	if logLevel == "" {
		logLevel = "info"
	}

	if err := gaywayLog.SetupLogger(logLevel, dev == "true"); err != nil {
		return fmt.Errorf("failed to set up logging: %w", err)
	}

	store, err := etcd.NewStore(
		strings.Split(etcdAddress, ","),
		etcdUsername,
		etcdPassword)
	if err != nil {
		return fmt.Errorf("failed to make etcd store: %w", err)
	}

	watermillLogger := zerowater.NewZerologLoggerAdapter(
		log.Logger.With().Str("component", "watermill").Logger(),
	)

	publisher, err := kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:   strings.Split(kafkaAddress, ","),
			Marshaler: kafka.DefaultMarshaler{},
		},
		watermillLogger,
	)
	if err != nil {
		return fmt.Errorf("failed to create kafka publisher: %w", err)
	}

	gatewayLogger := log.With().Str("component", "gateway").Logger()

	bot, err := gateway.NewGateway(gateway.Config{
		Token:     discordToken,
		Store:     store,
		Logger:    gatewayLogger,
		Publisher: publisher,
	})
	if err != nil {
		return fmt.Errorf("failed to start gateway: %w", err)
	}

	if err = bot.Start(); err != nil {
		return fmt.Errorf("failed to start gateway: %w", err)
	}

	if err = store.Close(); err != nil {
		return fmt.Errorf("failed to close etcd store: %w", err)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Error().Err(err).Msg("")
		os.Exit(1)
	}
}
