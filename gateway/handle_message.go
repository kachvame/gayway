package gateway

import "github.com/bwmarrin/discordgo"

func handleMessageEvents(msg *discordgo.Message) (topic Topic, key string) {
	return MessageEventTopic, msg.ID
}

func handleMessageCreate(_ *discordgo.Session, event *discordgo.MessageCreate) (topic Topic, key string) {
	return handleMessageEvents(event.Message)
}
