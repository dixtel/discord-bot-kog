package command

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/config"
	"github.com/dixtel/dicord-bot-kog/helpers"
	"github.com/dixtel/dicord-bot-kog/models"
	"github.com/rs/zerolog/log"
)

type DeclineCommand struct {
	applicationCommand *discordgo.ApplicationCommand
	Database           *models.Database
}

func (c *DeclineCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	user, err := c.Database.CreateOrGetUser(i.Member.User.Username, i.Member.User.ID)
	if err != nil {
		return fmt.Errorf("cannot create or get an user: %w", err)
	}

	isTestingChannel, err := c.Database.IsTestingChannel(i.ChannelID)
	if err != nil {
		return fmt.Errorf("cannot check if current channel is a testing channel: %w", err)
	}

	if !isTestingChannel {
		helpers.SendResponse(
			helpers.SendMessageTypeOriginator,
			"This is not a channel for testing",
			s,
			i,
		)

		return nil
	}

	canApproveMap, err := c.Database.UserCanApproveOrDeclineMap(user.ID)
	if err != nil {
		return fmt.Errorf("cannot check if user can approve/decline the map: %w", err)
	}

	if !canApproveMap {
		helpers.SendResponse(
			helpers.SendMessageTypeOriginator,
			"You cannot decline the map",
			s,
			i,
		)

		return nil
	}

	m, err := c.Database.GetLastUploadedMapByChannelID(i.ChannelID)
	if err != nil {
		return fmt.Errorf("cannot get last uploaded map by channel id: %w", err)
	}

	if m.Status != models.MapStatus_Testing {
		helpers.SendResponse(
			helpers.SendMessageTypeOriginator,
			fmt.Sprintf(
				"Cannot decline the map. Current map status is %q.",
				m.Status,
			),
			s,
			i,
		)
		return nil
	}

	data, err := c.Database.GetTestingChannelData(m.ID)
	if err != nil {
		return fmt.Errorf("cannot get testing channel data %w", err)
	}

	if _, ok := data.DeclinedBy[i.Member.User.ID]; ok {
		helpers.SendResponse(
			helpers.SendMessageTypeOriginator,
				"You already declined this map",
			s,
			i,
		)
		return nil
	}

	data.DeclinedBy[i.Member.User.ID] = struct{}{}

	err = c.Database.UpdateTestingChannelData(m.ID, data)
	if err != nil {
		return fmt.Errorf("cannot update testing channel data %w", err)
	}

	helpers.SendResponse(
		helpers.SendMessageTypeAll,
		// TODO: nick can be empty somehow?
		fmt.Sprintf("Map was declined by tester %s. (%v approvals / %v declines)", i.Member.Nick, len(data.ApprovedBy), len(data.DeclinedBy)),
		s,
		i,
	)

	return nil
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
