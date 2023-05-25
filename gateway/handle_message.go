package gateway

import "github.com/bwmarrin/discordgo"

func messageHandler(msg *discordgo.Message) (topic Topic, key string) {
	return MessageEventTopic, msg.ID
}

func messageReactionHandler(msg *discordgo.MessageReaction) (topic Topic, key string) {
	return MessageEventTopic, msg.MessageID
}

func MessageCreatePublisher(_ *discordgo.Session, event *discordgo.MessageCreate) (topic Topic, key string) {
	return messageHandler(event.Message)
}

func MessageUpdatePublisher(_ *discordgo.Session, event *discordgo.MessageUpdate) (topic Topic, key string) {
	return messageHandler(event.Message)
}

func MessageDeletePublisher(_ *discordgo.Session, event *discordgo.MessageDelete) (topic Topic, key string) {
	return messageHandler(event.Message)
}

func MessageReactionAddPublisher(_ *discordgo.Session, event *discordgo.MessageReactionAdd) (topic Topic, key string) {
	return messageReactionHandler(event.MessageReaction)
}

func MessageReactionRemovePublisher(_ *discordgo.Session, event *discordgo.MessageReactionRemove) (topic Topic, key string) {
	return messageReactionHandler(event.MessageReaction)
}

func MessageReactionRemoveAllPublisher(_ *discordgo.Session, event *discordgo.MessageReactionRemoveAll) (topic Topic, key string) {
	return messageReactionHandler(event.MessageReaction)
}
