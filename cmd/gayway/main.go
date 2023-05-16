package main

import (
	"fmt"
	"github.com/kachvame/gayway/gateway"
	"github.com/kachvame/gayway/kv/etcd"
	"log"
	"os"
	"strings"
)

func run() error {
	discordToken := os.Getenv("GAYWAY_DISCORD_TOKEN")
	etcdAddress := os.Getenv("GAYWAY_ETCD_ADDRESS")
	etcdPassword := os.Getenv("GAYWAY_ETCD_PASSWORD")
	etcdUsername := os.Getenv("GAYWAY_ETCD_USERNAME")

	store, err := etcd.NewStore(
		strings.Split(etcdAddress, ","),
		etcdUsername,
		etcdPassword)
	if err != nil {
		return fmt.Errorf("failed to make etcd store: %w", err)
	}

	bot, err := gateway.NewGateway(discordToken, store)
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
		log.Fatalln(err)
	}
}
