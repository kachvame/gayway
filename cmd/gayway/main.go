package main

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"os"
	"os/signal"
	"reflect"
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

	defer cli.Close()
	discord, err := discordgo.New(fmt.Sprintf("Bot %s", discordToken))
	if err != nil {
		return fmt.Errorf("failed to initialize discordgo: %w", err)
	}

	if sessionID != "" && sequence != 0 {
		setField(discord, "sequence", &sequence)
		setField(discord, "sessionID", sessionID)
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

	_, err = cli.Put(context.Background(), EtcdSequenceKey, strconv.Itoa(int(getSequence(discord))))
	if err != nil {
		return fmt.Errorf("failed to put sequence: %w", err)
	}

	_, err = cli.Put(context.Background(), EtcdSessionIDKey, getField(discord, "sessionID").String())
	if err != nil {
		return fmt.Errorf("failed to put session-id: %w", err)
	}

	return nil
}

func getField(s interface{}, fieldName string) reflect.Value {
	val := reflect.ValueOf(s)
	return reflect.Indirect(val).FieldByName(fieldName)
}

func getSequence(s *discordgo.Session) int64 {
	field := getField(s, "sequence")
	ptr := (*int64)(unsafe.Pointer(field.Pointer()))
	return atomic.LoadInt64(ptr)
}

func setField(s interface{}, fieldName string, newVal interface{}) {
	val := reflect.ValueOf(s)
	field := reflect.Indirect(val).FieldByName(fieldName)
	ptrToField := unsafe.Pointer(field.UnsafeAddr())
	settableField := reflect.NewAt(field.Type(), ptrToField).Elem()

	n := reflect.ValueOf(newVal)
	settableField.Set(n)
}

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}
