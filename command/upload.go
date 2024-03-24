package command

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/config"
	"github.com/dixtel/dicord-bot-kog/helpers"
	"github.com/dixtel/dicord-bot-kog/twmap"
	"github.com/rs/zerolog/log"
)

type UploadCommand struct {
	applicationCommand *discordgo.ApplicationCommand
}

func (c *UploadCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	attachment := getAttachment(i, "map-file")
	if attachment == nil {
		helpers.SendResponse(helpers.SendMessageTypeOriginator, "Cannot get attachment, please report this issue.", s, i)
		log.Error().Msg("cannot get attachment")
		return nil
	}

	// helpers.SendMessage(helpers.SendMessageTypeOriginator, fmt.Sprintf("got file %q", attachment.Filename), s, i)

	// attachment := i.Message.Attachments[0]
	// TODO: validate file

	// attachment.

	reader, err := twmap.MakeScreenshot(attachment.URL)
	if err != nil {
		return fmt.Errorf("cannot make screenshot of the map: %w", err)
	}
	defer reader.Close()

	helpers.SendResponseWithImage(
		helpers.SendMessageTypeAll,
		fmt.Sprintf("Screenshot of uploaded map %s", attachment.Filename),
		strings.Replace(attachment.Filename, ".map", ".png", 1),
		reader,
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
					Name:        "map-file",
					Description: "TODO",
					Required:    true,
				},
			},
		})
	if err != nil {
		log.Error().Err(err).Msgf("cannot application command: %q", c.GetName())
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
