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

type UploadCommand struct {
	applicationCommand *discordgo.ApplicationCommand
	Database           *models.Database
	BotRoles           *roles.BotRoles
}

func (c *UploadCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	var err error

	database, commitOrRollback := c.Database.Tx()
	defer commitOrRollback(&err)

	interaction, err := FromDiscordInteraction(s, i)
	if err != nil {
		return fmt.Errorf("cannot create interaction: %w", err)
	}

	if i.ChannelID != config.CONFIG.SubmitMapsChannelID {
		return interaction.SendMessage(
			fmt.Sprintf("This command can be used only in %s channel", config.CONFIG.SubmitMapsChannelName),
			InteractionMessageType_Private,
		)
	}

	user, err := database.CreateOrGetUser(i.Member.User.Username, i.Member.User.ID)
	if err != nil {
		return fmt.Errorf("cannot create or get an user: %w", err)
	}

	hasUnacceptedMap, err := database.UserHasUnacceptedMap(user.ID)
	if err != nil {
		return fmt.Errorf("cannot check if user have unaccepted map: %w", err)
	}

	if hasUnacceptedMap {
		return interaction.SendMessage(
			"You cannot upload a new map because you have another map waiting for acceptance.",
			InteractionMessageType_Private,
		)
	}

	attachment := getAttachmentFromOption(i, "file")
	if attachment == nil {
		return fmt.Errorf("cannot get attachment")
	}

	if !twmap.IsMapNameValid(attachment.Filename) {
		return interaction.SendMessage(
			"Incorrect map filename. Should be in format 'abc_1-2.map'",
			InteractionMessageType_Private,
		)
	}

	mapExists, err := database.MapExists(attachment.Filename)
	if err != nil {
		return fmt.Errorf("cannot check if map already exists: %w", err)
	}

	if mapExists {
		return interaction.SendMessage(
			"This map name is already taken. Please upload again with a different name",
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

	_, err = database.CreateMap(attachment.Filename, user.ID, mapSource, screenshotSource)
	if err != nil {
		return fmt.Errorf("cannot create map %w", err)
	}

	return interaction.SendMessageWithPNGImage(
		fmt.Sprintf(
			"New map %s from %s!", 
			strings.Replace(attachment.Filename, ".map", "", 1), 
			mentionUser(i),
		),
		InteractionMessageType_Public,
		strings.Replace(attachment.Filename, ".map", ".png", 1),
		screenshotSource,
	)
}

func (c *UploadCommand) GetName() string {
	return "upload"
}

func (c *UploadCommand) GetDescription() string {
	return "Upload a new map"
}

func (c *UploadCommand) ApplicationCommandCreate(s *discordgo.Session) {
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
					Description: "Map file",
					Required:    true,
				},
			},
		})
	if err != nil {
		log.Error().Err(err).Msgf("cannot create application command: %q", c.GetName())
	}

	c.applicationCommand = applicationCommand
}

func (c *UploadCommand) ApplicationCommandDelete(s *discordgo.Session) {
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
