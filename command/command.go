package command

import "github.com/bwmarrin/discordgo"

type CommandInterface interface {
	Handle(s *discordgo.Session, i *discordgo.InteractionCreate) error
	GetName() string
	GetDescription() string
	ApplicationCommandCreate(s *discordgo.Session) error
	ApplicationCommandDelete(s *discordgo.Session)
}