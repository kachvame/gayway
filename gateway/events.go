package gateway

import "github.com/bwmarrin/discordgo"

type Topic string

const (
	MessageEventTopic Topic = "message"
)

type Handler interface {
	Handle(session *discordgo.Session, event *discordgo.Event) (topic Topic, key string)
}

type HandlerFunc[T any] func(session *discordgo.Session, event T) (topic Topic, key string)

func (e HandlerFunc[T]) Handle(session *discordgo.Session, rawEvent *discordgo.Event) (topic Topic, key string) {
	event, ok := rawEvent.Struct.(T)
	if !ok {
		return
	}
	return e(session, event)
}

func handler[T any](fn func(session *discordgo.Session, event T) (topic Topic, key string)) HandlerFunc[T] {
	return fn
}
