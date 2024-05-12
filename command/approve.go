package command

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/config"
	"github.com/dixtel/dicord-bot-kog/models"
	"github.com/dixtel/dicord-bot-kog/roles"
	"github.com/rs/zerolog/log"
)

type ApproveCommand struct {
	applicationCommand *discordgo.ApplicationCommand
	Database           *models.Database
	BotRoles           *roles.BotRoles
}

func (c *ApproveCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	var err error

	database, commitOrRollback := c.Database.Tx()
	defer commitOrRollback(&err)

	err = approveOrDecline(database, c.BotRoles, s, i, ApproveORDecline_Approve)

	return err
}

func (c *ApproveCommand) GetName() string {
	return "approve"
}

func (c *ApproveCommand) GetDescription() string {
	return "Approve the map"
}

func (c *ApproveCommand) ApplicationCommandCreate(s *discordgo.Session) error {
	applicationCommand, err := s.ApplicationCommandCreate(
		config.CONFIG.AppID,
		config.CONFIG.GuildID,
		&discordgo.ApplicationCommand{
			Name:        c.GetName(),
			Type:        discordgo.ChatApplicationCommand,
			Description: c.GetDescription(),
		})
	if err != nil {
		return fmt.Errorf("cannot create application command: %q", c.GetName())
	}

	c.applicationCommand = applicationCommand

	return nil
}

func (c *ApproveCommand) ApplicationCommandDelete(s *discordgo.Session) {
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
