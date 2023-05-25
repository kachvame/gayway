package gateway

import "github.com/bwmarrin/discordgo"

type Topic string

const (
	MessageEventTopic Topic = "messages"
)

type Handler interface {
	Handle(session *discordgo.Session, event *discordgo.Event) (topic Topic, key string, ok bool)
}

type HandlerFunc[T any] func(session *discordgo.Session, event T) (topic Topic, key string)

func (h HandlerFunc[T]) Handle(session *discordgo.Session, rawEvent *discordgo.Event) (Topic, string, bool) {
	event, ok := rawEvent.Struct.(T)
	if !ok {
		return "", "", false
	}
	topic, key := h(session, event)
	return topic, key, true
}

func handler[T any](fn func(session *discordgo.Session, event T) (topic Topic, key string)) HandlerFunc[T] {
	return fn
}
