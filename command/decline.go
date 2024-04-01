package command

import (
	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/config"
	"github.com/dixtel/dicord-bot-kog/models"
	"github.com/dixtel/dicord-bot-kog/roles"
	"github.com/rs/zerolog/log"
)

type DeclineCommand struct {
	applicationCommand *discordgo.ApplicationCommand
	Database           *models.Database
	BotRoles           *roles.BotRoles
}

func (c *DeclineCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	var err error

	database, commitOrRollback := c.Database.Tx()
	defer commitOrRollback(&err)

	err = approveOrDecline(database, c.BotRoles, s, i, ApproveORDecline_Decline)

	return err
}

func (c *DeclineCommand) GetName() string {
	return "decline"
}

func (c *DeclineCommand) GetDescription() string {
	return "Decline the map"
}

func (c *DeclineCommand) ApplicationCommandCreate(s *discordgo.Session) {
	applicationCommand, err := s.ApplicationCommandCreate(
		config.CONFIG.AppID,
		config.CONFIG.GuildID,
		&discordgo.ApplicationCommand{
			Name:        c.GetName(),
			Type:        discordgo.ChatApplicationCommand,
			Description: c.GetDescription(),
		})
	if err != nil {
		log.Error().Err(err).Msgf("cannot create application command: %q", c.GetName())
	}

	c.applicationCommand = applicationCommand
}

func (c *DeclineCommand) ApplicationCommandDelete(s *discordgo.Session) {
	log := log.With().Str("command-name", c.GetName()).Logger()

	if c.applicationCommand == nil {
		log.Error().Msgf("cannot delete application command: application command is nil")
		return
	}

	err := s.ApplicationCommandDelete(config.CONFIG.AppID, config.CONFIG.GuildID, c.applicationCommand.ID)
	if err != nil {
		log.Error().Err(err).Msgf("cannot delete application command")
	}
}
