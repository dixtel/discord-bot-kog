package command

import (
	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/helpers"
)

func getAttachmentFromOption(i *discordgo.InteractionCreate, optionName string) *discordgo.MessageAttachment {
	option := helpers.First(
		i.ApplicationCommandData().Options,
		func(val *discordgo.ApplicationCommandInteractionDataOption) bool {
			return val.Name == optionName
		},
	)
	if option == nil {
		return nil
	}

	// TODO: prevent from panic
	return i.ApplicationCommandData().Resolved.Attachments[(*option).Value.(string)]
}

func getUserFromOption(s *discordgo.Session, i *discordgo.InteractionCreate, optionName string) *discordgo.User {
	option := helpers.First(
		i.ApplicationCommandData().Options,
		func(val *discordgo.ApplicationCommandInteractionDataOption) bool {
			return val.Name == optionName
		},
	)
	if option == nil {
		return nil
	}

	return (*option).UserValue(s)
}

func getUsername(i *discordgo.InteractionCreate) string {
	if i.Member != nil && i.Member.Nick != ""  {
		return i.Member.Nick
	} 

	if i.Member.User != nil && i.Member.User.Username != ""  {
		return i.Member.User.Username
	} 

	return "<undefined>"
}
