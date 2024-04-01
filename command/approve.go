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
	interaction, err := FromDiscordInteraction(s, i)
	if err != nil {
		return fmt.Errorf("cannot create interaction: %w", err)
	}

	_, err = c.Database.CreateOrGetUser(i.Member.User.Username, i.Member.User.ID)
	if err != nil {
		return fmt.Errorf("cannot create or get an user: %w", err)
	}

	isTestingChannel, err := c.Database.IsTestingChannel(i.ChannelID)
	if err != nil {
		return fmt.Errorf("cannot check if current channel is a testing channel: %w", err)
	}

	if !isTestingChannel {
		return interaction.SendMessage(
			"This is not a channel for testing",
			InteractionMessageType_Private,
		)
	}

	if !c.BotRoles.HasMapAcceptorRole(i.Member) && !c.BotRoles.HasMapTesterRole(i.Member) {
		return interaction.SendMessage(
			"You don't have role 'Tester' or 'Map Acceptor' to approve the map.",
			InteractionMessageType_Private,
		)
	}

	m, err := c.Database.GetLastUploadedMapByChannelID(i.ChannelID)
	if err != nil {
		return fmt.Errorf("cannot get last uploaded map by channel id: %w", err)
	}

	if m.Status != models.MapStatus_Testing {
		return interaction.SendMessage(
			fmt.Sprintf("Cannot approve the map. Current map status is %q.", m.Status),
			InteractionMessageType_Private,
		)
	}

	data, err := c.Database.GetTestingChannelData(m.ID)
	if err != nil {
		return fmt.Errorf("cannot get testing channel data %w", err)
	}

	if _, ok := data.ApprovedBy[i.Member.User.ID]; ok {
		return interaction.SendMessage(
			"You already approved this map",
			InteractionMessageType_Private,
		)
	}

	data.ApprovedBy[i.Member.User.ID] = struct{}{}

	err = c.Database.UpdateTestingChannelData(m.ID, data)
	if err != nil {
		return fmt.Errorf("cannot update testing channel data %w", err)
	}

	return interaction.SendMessage(
		fmt.Sprintf(
			"Map was approved by tester %s. (%v approvals / %v declines)",
			getUsername(i), len(data.ApprovedBy), len(data.DeclinedBy),
		),
		InteractionMessageType_Public,
	)
}

func (c *ApproveCommand) GetName() string {
	return "approve"
}

func (c *ApproveCommand) GetDescription() string {
	return "Approve the map"
}

func (c *ApproveCommand) ApplicationCommandCreate(s *discordgo.Session) {
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
