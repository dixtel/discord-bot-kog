package command

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/channel"
	"github.com/dixtel/dicord-bot-kog/config"
	"github.com/dixtel/dicord-bot-kog/helpers"
	"github.com/dixtel/dicord-bot-kog/models"
	"github.com/dixtel/dicord-bot-kog/roles"
	"github.com/dixtel/dicord-bot-kog/twmap"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type AcceptCommand struct {
	applicationCommand  *discordgo.ApplicationCommand
	Database            *models.Database
	BotRoles            *roles.BotRoles
	SubmitMapsChannelID string
}

func (c *AcceptCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	var err error

	database, commitOrRollback := c.Database.Tx()
	defer commitOrRollback(&err)

	interaction, err := FromDiscordInteraction(s, i)
	if err != nil {
		return fmt.Errorf("cannot create interaction: %w", err)
	}

	if i.ChannelID != c.SubmitMapsChannelID {
		return interaction.SendMessage(
			fmt.Sprintf("This command can be used only in %s channel", config.CONFIG.SubmitMapsChannelName),
			InteractionMessageType_Private,
		)
	}

	_, err = database.CreateOrGetUser(i.Member.User.Username, i.Member.User.ID)
	if err != nil {
		return fmt.Errorf("cannot create or get an user: %w", err)
	}

	if !c.BotRoles.HasMapAcceptorRole(i.Member) {
		return interaction.SendMessage(
			"You don't have role 'Map Acceptor' to accept the map.",
			InteractionMessageType_Private,
		)
	}

	mapCreator := getUserFromOption(s, i, "user")

	m, err := database.GetLastUploadedMap(mapCreator.ID)

	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return interaction.SendMessage(
			"This user doesn't have any uploaded map waiting to be accepted",
			InteractionMessageType_Private,
		)
	} else if err != nil {
		return fmt.Errorf("cannot get last uploaded map: %w", err)
	}

	if m.Status != models.MapStatus_WaitingToAccept {
		return interaction.SendMessage(
			fmt.Sprintf(
				"This user doesn't have any uploaded map waiting to be accepted or rejected. The last map status for this user is %q.",
				m.Status,
			),
			InteractionMessageType_Private,
		)
	}

	channelName := fmt.Sprintf(
		config.CONFIG.TestingChannelFormat,
		twmap.RemoveMapFileExtension(m.FileName),
	)

	testingMapChannel, err := channel.CreateTestingMapChannel(s, channelName, c.BotRoles, mapCreator.ID)
	if err != nil {
		return fmt.Errorf("cannot create testing map channel: %w", err)
	}

	testingChannel, err := database.CreateTestingChannel(testingMapChannel.GetID(), channelName)
	if err != nil {
		return fmt.Errorf("cannot create testing channel record: %w", err)
	}

	err = database.AcceptMap(m.ID, mapCreator.ID, testingChannel.ID)
	if err != nil {
		_, e := s.ChannelDelete(testingMapChannel.GetID())
		if e != nil {
			log.Error().Err(err).Msg("cannot delete discord channel")
		}

		e = database.DeleteTestingChannel(testingMapChannel.GetID())
		if e != nil {
			log.Error().Err(err).Msg("cannot delete testing channel record")
		}

		return fmt.Errorf("cannot mark map as accepted: %w", err)
	}

	_, err = s.ChannelMessageSendComplex(
		testingMapChannel.GetID(),
		&discordgo.MessageSend{
			Content: fmt.Sprintf(
				"New map %s from %s!\n%s and %s can now /approve or /decline the map and discuss about details with the author.",
				twmap.RemoveMapFileExtension(m.FileName),
				mentionUser(i),
				c.BotRoles.Mention(c.BotRoles.MapAcceptor),
				c.BotRoles.Mention(c.BotRoles.MapTester),
			),
			Files: []*discordgo.File{
				{
					Name:        strings.Replace(m.FileName, ".map", ".png", 1),
					ContentType: "image/png",
					Reader:      bytes.NewReader(m.Screenshot),
				},
				{
					Name:        m.FileName,
					ContentType: "text/plain",
					Reader:      bytes.NewReader(m.File),
				},
			},
		},
	)
	if err != nil {
		return fmt.Errorf("cannot send message to channel: %w", err)
	}

	return interaction.SendMessage(
		fmt.Sprintf("Map %s from %s was accepted for the next stage - testing 🎉",
			twmap.RemoveMapFileExtension(m.FileName), mentionUser(i)),
		InteractionMessageType_Public,
	)
}

func (c *AcceptCommand) GetName() string {
	return "accept"
}

func (c *AcceptCommand) GetDescription() string {
	return "Accept the map. This command will create a new testing channel"
}

func (c *AcceptCommand) ApplicationCommandCreate(s *discordgo.Session) error {
	applicationCommand, err := s.ApplicationCommandCreate(
		config.CONFIG.AppID,
		config.CONFIG.GuildID,
		&discordgo.ApplicationCommand{
			Name:                     c.GetName(),
			Type:                     discordgo.ChatApplicationCommand,
			Description:              c.GetDescription(),
			DefaultMemberPermissions: helpers.ToPtr(int64(0)),
			Options: []*discordgo.ApplicationCommandOption{
				{

					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "Accept the user map",
					Required:    true,
				},
			},
		})
	if err != nil {
		return fmt.Errorf("cannot create application command: %q", c.GetName())
	}

	c.applicationCommand = applicationCommand

	return nil
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
