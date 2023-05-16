package main

import (
	"fmt"
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
	logLevel := os.Getenv("GAYWAY_LOG_LEVEL")
	dev := os.Getenv("GAYWAY_DEV")

	if logLevel == "" {
		logLevel = "info"
	}

	fmt.Println(dev)

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

	gatewayLogger := log.With().Str("component", "gateway").Logger()

	bot, err := gateway.NewGateway(gateway.Config{
		Token:  discordToken,
		Store:  store,
		Logger: gatewayLogger,
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
