package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/websocket"
	"github.com/kachvame/gayway/kv"
	"github.com/kachvame/gayway/reflection"
	"github.com/rs/zerolog"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"unsafe"
)

const (
	SequenceKey  = "gayway/sequence"
	SessionIDKey = "gayway/session-id"
)

type Config struct {
	Token     string
	Store     kv.Store
	Logger    zerolog.Logger
	Publisher message.Publisher
}

type Gateway struct {
	session   *discordgo.Session
	store     kv.Store
	logger    zerolog.Logger
	publisher message.Publisher
}

func NewGateway(config Config) (*Gateway, error) {
	session, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize discord session: %w", err)
	}

	gateway := &Gateway{
		session:   session,
		store:     config.Store,
		logger:    config.Logger,
		publisher: config.Publisher,
	}

	publishingHandlers := []Handler{
		handler(MessageCreatePublisher),
		handler(MessageUpdatePublisher),
		handler(MessageDeletePublisher),
		handler(MessageReactionAddPublisher),
		handler(MessageReactionRemovePublisher),
		handler(MessageReactionRemoveAllPublisher),
	}

	for _, publishingHandler := range publishingHandlers {
		session.AddHandler(
			gateway.CreatePublishingHandler(
				publishingHandler,
			),
		)
	}

	session.AddHandler(gateway.Ready)
	session.AddHandler(gateway.Resumed)

	return gateway, nil
}

func (gateway *Gateway) Start() error {
	sequence := int64(0)
	sequenceBytes, err := gateway.store.Get(context.Background(), SequenceKey)
	if err != nil && err != kv.ErrNotFound {
		return fmt.Errorf("failed to get sequence from store: %w", err)
	}

	if sequenceBytes != nil {
		sequence, err = strconv.ParseInt(string(sequenceBytes), 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse sequence: %w", err)
		}
	}

	sessionID, err := gateway.store.Get(context.Background(), SessionIDKey)
	if err != nil && err != kv.ErrNotFound {
		return fmt.Errorf("failed to get session id from store: %w", err)
	}

	reflection.SetField(gateway.session, "sequence", &sequence)
	reflection.SetField(gateway.session, "sessionID", string(sessionID))

	if err = gateway.session.Open(); err != nil {
		return fmt.Errorf("failed to open discord session: %w", err)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	if err = gateway.session.CloseWithCode(websocket.CloseServiceRestart); err != nil {
		return fmt.Errorf("failed to close discord session: %w", err)
	}

	sequenceField := reflection.GetField(gateway.session, "sequence")
	sequenceAddress := (*int64)(unsafe.Pointer(sequenceField.Pointer()))
	sequence = atomic.LoadInt64(sequenceAddress)

	sequenceValue := []byte(strconv.FormatInt(sequence, 10))
	if err = gateway.store.Set(context.Background(), SequenceKey, sequenceValue); err != nil {
		return fmt.Errorf("failed to put sequence: %w", err)
	}

	sessionID = []byte(reflection.GetField(gateway.session, "sessionID").String())
	if err = gateway.store.Set(context.Background(), SessionIDKey, sessionID); err != nil {
		return fmt.Errorf("failed to put session id: %w", err)
	}

	return nil
}

func (gateway *Gateway) Ready(_ *discordgo.Session, _ *discordgo.Ready) {
	gateway.logger.
		Info().
		Msg("Received ready")
}

func (gateway *Gateway) Resumed(_ *discordgo.Session, _ *discordgo.Resumed) {
	gateway.logger.
		Info().
		Msg("Received resumed")
}

func (gateway *Gateway) CreatePublishingHandler(handler Handler) func(s *discordgo.Session, event *discordgo.Event) {
	return func(session *discordgo.Session, event *discordgo.Event) {
		topic, key, ok := handler.Handle(session, event)
		if !ok {
			return
		}

		payload, err := json.Marshal(Message{
			Event: event.RawData,
			Type:  event.Type,
		})
		if err != nil {
			gateway.logger.Error().Err(err).Msg("Error occurred during message marshalling")
			return
		}

		msg := message.NewMessage(
			watermill.NewULID(),
			payload,
		)

		msg.Metadata.Set("partition", key)

		if err := gateway.publisher.Publish(fmt.Sprintf("gayway-%s", topic), msg); err != nil {
			gateway.logger.Error().Err(err).Msg("Error occurred during publishing")
			return
		}
	}
}
