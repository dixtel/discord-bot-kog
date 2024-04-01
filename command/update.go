package command

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/config"
	"github.com/dixtel/dicord-bot-kog/models"
	"github.com/dixtel/dicord-bot-kog/roles"
	"github.com/dixtel/dicord-bot-kog/twmap"
	"github.com/rs/zerolog/log"
)

type UpdateCommand struct {
	applicationCommand *discordgo.ApplicationCommand
	Database           *models.Database
	BotRoles           *roles.BotRoles
}

func (c *UpdateCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	interaction, err := FromDiscordInteraction(s, i)
	if err != nil {
		return fmt.Errorf("cannot create interaction: %w", err)
	}

	user, err := c.Database.CreateOrGetUser(i.Member.User.Username, i.Member.User.ID)
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

	canUpdateMap, err := c.Database.UserCanUpdateMap(user.ID, i.ChannelID)
	if err != nil {
		return fmt.Errorf("cannot check if user can update the map: %w", err)
	}

	if !canUpdateMap {
		return interaction.SendMessage(
			"You cannot update the map",
			InteractionMessageType_Private,
		)
	}

	attachment := getAttachmentFromOption(i, "file")
	if attachment == nil {
		return fmt.Errorf("cannot get attachment")
	}

	m, err := c.Database.GetLastUploadedMap(i.Member.User.ID)
	if err != nil {
		return fmt.Errorf("cannot get last uploaded map: %w", err)
	}

	if m.Status != models.MapStatus_Testing {
		return interaction.SendMessage(
			fmt.Sprintf("Cannot update the map. Current map status is %q.", m.Status),
			InteractionMessageType_Private,
		)
	}

	if m.Name != attachment.Filename {
		return interaction.SendMessage(
			fmt.Sprintf("Your map file should be named %q instead of %q", m.Name, attachment.Filename),
			InteractionMessageType_Private,
		)
	}

	mapSource, err := twmap.DownloadMapFromDiscord(attachment.URL)
	if err != nil {
		return fmt.Errorf("cannot download map from discord %w", err)
	}

	screenshotSource, err := twmap.MakeScreenshot(mapSource)
	if err != nil {
		return fmt.Errorf("cannot make screenshot of the map: %w", err)
	}

	err = c.Database.UpdateMap(m.ID, mapSource, screenshotSource)
	if err != nil {
		return fmt.Errorf("cannot create map %w", err)
	}

	err = c.Database.UpdateTestingChannelData(m.ID, &models.TestingChannelData{
		ApprovedBy: map[string]struct{}{},
		DeclinedBy: map[string]struct{}{},
	})
	if err != nil {
		return fmt.Errorf("cannot update testing channel data %w", err)
	}

	// TODO: to show map diff we need to save previous screenshot in database


	return interaction.SendMessageWithPNGImage(
		fmt.Sprintf("Map %s was updated by %s", strings.Replace(attachment.Filename, ".map", "", 1), mentionUser(i)),
		InteractionMessageType_Private,
		strings.Replace(attachment.Filename, ".map", ".png", 1),
		screenshotSource,
	)
}

func (c *UpdateCommand) GetName() string {
	return "update"
}

func (c *UpdateCommand) GetDescription() string {
	return "Submit an updated map"
}

func (c *UpdateCommand) ApplicationCommandCreate(s *discordgo.Session) {
	applicationCommand, err := s.ApplicationCommandCreate(
		config.CONFIG.AppID,
		config.CONFIG.GuildID,
		&discordgo.ApplicationCommand{
			Name:        c.GetName(),
			Type:        discordgo.ChatApplicationCommand,
			Description: c.GetDescription(),

			Options: []*discordgo.ApplicationCommandOption{
				{

					Type:        discordgo.ApplicationCommandOptionAttachment,
					Name:        "file",
					Description: "Updated map file",
					Required:    true,
				},
			},
		})
	if err != nil {
		log.Error().Err(err).Msgf("cannot create application command: %q", c.GetName())
	}

	c.applicationCommand = applicationCommand
}

func (c *UpdateCommand) ApplicationCommandDelete(s *discordgo.Session) {
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
