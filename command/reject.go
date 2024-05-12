package command

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/config"
	"github.com/dixtel/dicord-bot-kog/helpers"
	"github.com/dixtel/dicord-bot-kog/models"
	"github.com/dixtel/dicord-bot-kog/roles"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type RejectCommand struct {
	applicationCommand  *discordgo.ApplicationCommand
	Database            *models.Database
	BotRoles            *roles.BotRoles
	SubmitMapsChannelID string
}

func (c *RejectCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
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
			"You don't have role 'Map Acceptor' to reject the map.",
			InteractionMessageType_Private,
		)
	}

	mapCreator := getUserFromOption(s, i, "user")

	m, err := database.GetLastUploadedMap(mapCreator.ID)

	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return interaction.SendMessage(
			"This user doesn't have any uploaded map waiting to be rejected",
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

	err = database.RejectMap(m.ID, mapCreator.ID, m.ID)
	if err != nil {
		return fmt.Errorf("cannot mark map as rejected: %w", err)
	}

	return interaction.SendMessage(
		"The map was rejected",
		InteractionMessageType_Private,
	)
}

func (c *RejectCommand) GetName() string {
	return "reject"
}

func (c *RejectCommand) GetDescription() string {
	return "Reject the map."
}

func (c *RejectCommand) ApplicationCommandCreate(s *discordgo.Session) error {
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
					Description: "Reject the user map",
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

func (c *RejectCommand) ApplicationCommandDelete(s *discordgo.Session) {
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
