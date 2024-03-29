package command

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/config"
	"github.com/dixtel/dicord-bot-kog/helpers"
	"github.com/dixtel/dicord-bot-kog/models"
	"github.com/dixtel/dicord-bot-kog/twmap"
	"github.com/rs/zerolog/log"
)

type UploadCommand struct {
	applicationCommand *discordgo.ApplicationCommand
	Database           *models.Database
}

func (c *UploadCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if i.ChannelID != config.CONFIG.SubmitMapsChannelID {
		helpers.SendResponse(
			helpers.SendMessageTypeOriginator,
			fmt.Sprintf("This command can be used only in %s channel", config.CONFIG.SubmitMapsChannelName),
			s,
			i,
		)

		return nil
	}

	user, err := c.Database.CreateOrGetUser(i.Member.User.Username, i.Member.User.ID)
	if err != nil {
		return fmt.Errorf("cannot create or get an user: %w", err)
	}

	hasUnacceptedMap, err := c.Database.UserHasUnacceptedMap(user.ID)
	if err != nil {
		return fmt.Errorf("cannot check if user have unaccepted map: %w", err)
	}

	if hasUnacceptedMap {
		helpers.SendResponse(
			helpers.SendMessageTypeOriginator,
			"You cannot upload a new map because you have another map waiting for acceptance.",
			s,
			i,
		)

		return nil
	}

	attachment := getAttachmentFromOption(i, "file")
	if attachment == nil {
		return fmt.Errorf("cannot get attachment")
	}

	// TODO: map name validation

	mapExists, err := c.Database.MapExists(attachment.Filename)
	if err != nil {
		return fmt.Errorf("cannot check if map already exists: %w", err)
	}

	if mapExists {
		helpers.SendResponse(
			helpers.SendMessageTypeOriginator,
			"This map name is already taken. Please upload again with a different name",
			s,
			i,
		)

		return nil
	}

	mapSource, err := twmap.DownloadMapFromDiscord(attachment.URL)
	if err != nil {
		return fmt.Errorf("cannot download map from discord %w", err)
	}

	screenshotSource, err := twmap.MakeScreenshot(mapSource)
	if err != nil {
		return fmt.Errorf("cannot make screenshot of the map: %w", err)
	}

	_, err = c.Database.CreateMap(attachment.Filename, user.ID, mapSource)
	if err != nil {
		return fmt.Errorf("cannot create map %w", err)
	}

	helpers.SendResponseWithImage(
		helpers.SendMessageTypeAll,
		// TODO: i.Member.Nick is empty
		fmt.Sprintf("Screenshot of uploaded map %s by %s", attachment.Filename, i.Member.Nick),
		strings.Replace(attachment.Filename, ".map", ".png", 1),
		bytes.NewReader(screenshotSource),
		s,
		i,
	)

	return nil
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
