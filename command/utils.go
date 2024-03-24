package command

import (
	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/helpers"
)

func getAttachment(i *discordgo.InteractionCreate, optionName string) *discordgo.MessageAttachment {
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
