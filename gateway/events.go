package gateway

import "github.com/bwmarrin/discordgo"

type EventTopic string

const (
	MessageEventTopic EventTopic = "message"
)

type EventHandlerFunc[T any] func(session *discordgo.Session, event T) (topic EventTopic, key string)

func handleMessageEvents(msg *discordgo.Message) (topic EventTopic, key string) {
	return MessageEventTopic, msg.ID
}
