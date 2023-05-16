package main

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/websocket"
	"github.com/kachvame/gayway/reflection"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"
)

const EtcdSequenceKey = "gayway/sequence"
const EtcdSessionIDKey = "gayway/session-id"

func run() error {
	discordToken := os.Getenv("GAYWAY_DISCORD_TOKEN")
	etcdAddress := os.Getenv("GAYWAY_ETCD_ADDRESS")
	etcdPassword := os.Getenv("GAYWAY_ETCD_PASSWORD")
	etcdUsername := os.Getenv("GAYWAY_ETCD_USERNAME")

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(etcdAddress, ","),
		Username:    etcdUsername,
		Password:    etcdPassword,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("failed to dial etcd: %w", err)
	}

	resp, err := cli.Get(context.Background(), EtcdSessionIDKey)
	if err != nil {
		return fmt.Errorf("failed to dial etcd: %w", err)
	}

	var sessionID string
	for _, kv := range resp.Kvs {
		sessionID = string(kv.Value)
	}

	resp, err = cli.Get(context.Background(), EtcdSequenceKey)
	if err != nil {
		return fmt.Errorf("failed to dial etcd: %w", err)
	}

	var sequence int64
	for _, kv := range resp.Kvs {
		seq, err := strconv.Atoi(string(kv.Value))
		if err != nil {
			return fmt.Errorf("failed to parse sequence: %w", err)
		}
		sequence = int64(seq)
	}

	defer func(cli *clientv3.Client) {
		_ = cli.Close()
	}(cli)

	discord, err := discordgo.New(fmt.Sprintf("Bot %s", discordToken))
	if err != nil {
		return fmt.Errorf("failed to initialize discordgo: %w", err)
	}

	if sessionID != "" && sequence != 0 {
		reflection.SetField(discord, "sequence", &sequence)
		reflection.SetField(discord, "sessionID", sessionID)
	}

	discord.AddHandler(func(se *discordgo.Session, e *discordgo.Event) {
		fmt.Println(e.Type)
	})

	if err = discord.Open(); err != nil {
		return fmt.Errorf("failed to open discord client: %w", err)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	if err := discord.CloseWithCode(websocket.CloseServiceRestart); err != nil {
		return fmt.Errorf("failed to close discord conn: %w", err)
	}

	_, err = cli.Put(context.Background(), EtcdSequenceKey, strconv.Itoa(int(getSequence(discord))))
	if err != nil {
		return fmt.Errorf("failed to put sequence: %w", err)
	}

	_, err = cli.Put(context.Background(), EtcdSessionIDKey, reflection.GetField(discord, "sessionID").String())
	if err != nil {
		return fmt.Errorf("failed to put session-id: %w", err)
	}

	return nil
}

func getSequence(s *discordgo.Session) int64 {
	field := reflection.GetField(s, "sequence")
	ptr := (*int64)(unsafe.Pointer(field.Pointer()))
	return atomic.LoadInt64(ptr)
}

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}
