package command

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/config"
	"github.com/dixtel/dicord-bot-kog/helpers"
	"github.com/dixtel/dicord-bot-kog/models"
	"github.com/dixtel/dicord-bot-kog/twmap"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type AcceptCommand struct {
	applicationCommand *discordgo.ApplicationCommand
	Database           *models.Database
}

func (c *AcceptCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if i.ChannelID != config.CONFIG.SubmitMapsChannelID {
		helpers.SendResponse(
			helpers.SendMessageTypeOriginator,
			fmt.Sprintf("This command can be used only in %s channel", config.CONFIG.SubmitMapsChannelName),
			s,
			i,
		)

		return nil
	}

	invocator, err := c.Database.CreateOrGetUser(i.Member.User.Username, i.Member.User.ID)
	if err != nil {
		return fmt.Errorf("cannot create or get an user: %w", err)
	}

	if !invocator.HasRole(models.RoleName_MapAcceptor) {
		helpers.SendResponse(
			helpers.SendMessageTypeOriginator,
			"You don't have permission to accept the map.",
			s,
			i,
		)

		return nil
	}

	mapCreator := getUserFromOption(s, i, "user")

	m, err := c.Database.GetLastUploadedMap(mapCreator.ID)

	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		helpers.SendResponse(
			helpers.SendMessageTypeOriginator,
			"This user doesn't have any uploaded map waiting to be accepted",
			s,
			i,
		)

		return nil
	} else if err != nil {
		return fmt.Errorf("cannot get last uploaded map: %w", err)
	}

	if m.Status != models.MapStatus_WaitingToAccept {
		helpers.SendResponse(
			helpers.SendMessageTypeOriginator,
			fmt.Sprintf(
				"This user doesn't have any uploaded map waiting to be accepted. The last map status for this user is %q.",
				m.Status,
			),
			s,
			i,
		)
		return nil
	}

	channelName := fmt.Sprintf(config.CONFIG.TestingChannelFormat, twmap.FixMapName(m.Name))

	// TODO: make channel private; invite testers, map acceptors and map creator
	discordChannel, err := s.GuildChannelCreate(
		config.CONFIG.GuildID,
		channelName,
		discordgo.ChannelTypeGuildText,
	)
	if err != nil {
		return fmt.Errorf("cannot create discord channel: %w", err)
	}

	testingChannel, err := c.Database.CreateTestingChannel(discordChannel.ID, channelName)
	if err != nil {
		return fmt.Errorf("cannot create testing channel record: %w", err)
	}

	err = c.Database.AcceptMap(m.ID, mapCreator.ID, testingChannel.ID)
	if err != nil {
		_, e := s.ChannelDelete(discordChannel.ID)
		if e != nil {
			log.Error().Err(err).Msg("cannot delete discord channel")
		}

		e = c.Database.DeleteTestingChannel(discordChannel.ID)
		if e != nil {
			log.Error().Err(err).Msg("cannot delete testing channel record")
		}

		return fmt.Errorf("cannot mark map as accepted: %w", err)
	}

	helpers.SendResponse(
		helpers.SendMessageTypeOriginator,
		fmt.Sprintf(
			// "The map was accepted. A new channel mapping_channel_BMO.map was created."
			"The map was accepted. A new channel %s was created.",
			channelName,
		),
		s,
		i,
	)

	return nil
}

func (c *AcceptCommand) GetName() string {
	return "accept"
}

func (c *AcceptCommand) GetDescription() string {
	return "Accept the map. This command will create a new testing channel"
}

func (c *AcceptCommand) ApplicationCommandCreate(s *discordgo.Session) {
	applicationCommand, err := s.ApplicationCommandCreate(
		config.CONFIG.AppID,
		config.CONFIG.GuildID,
		&discordgo.ApplicationCommand{
			Name:        c.GetName(),
			Type:        discordgo.ChatApplicationCommand,
			Description: c.GetDescription(),

			Options: []*discordgo.ApplicationCommandOption{
				{

					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "Accept the user map",
					Required:    true,
				},
			},
		})
	// TODO: this  error handling can be moved one level higher, just return the error here
	if err != nil {
		log.Error().Err(err).Msgf("cannot create application command: %q", c.GetName())
	}

	c.applicationCommand = applicationCommand
}

func (c *AcceptCommand) ApplicationCommandDelete(s *discordgo.Session) {
	log := log.With().Str("command-name", c.GetName()).Logger()

	// TODO: the errors here can be handled one level higher

	if c.applicationCommand == nil {
		log.Error().Msgf("cannot delete application command: application command is nil")
		return
	}

	err := s.ApplicationCommandDelete(config.CONFIG.AppID, config.CONFIG.GuildID, c.applicationCommand.ID)
	if err != nil {
		log.Error().Err(err).Msgf("cannot delete application command")
	}
}
