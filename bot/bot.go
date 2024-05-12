package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/channel"
	"github.com/dixtel/dicord-bot-kog/command"
	"github.com/dixtel/dicord-bot-kog/config"
	"github.com/dixtel/dicord-bot-kog/helpers"
	"github.com/dixtel/dicord-bot-kog/models"
	"github.com/dixtel/dicord-bot-kog/roles"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func createCommands(
	db *models.Database,
	roles *roles.BotRoles,
	submitMapsChannelID string,
) []command.CommandInterface {
	return []command.CommandInterface{
		&command.UploadCommand{
			Database:            db,
			BotRoles:            roles,
			SubmitMapsChannelID: submitMapsChannelID,
		},
		&command.AcceptCommand{
			Database:            db,
			BotRoles:            roles,
			SubmitMapsChannelID: submitMapsChannelID,
		},
		&command.RejectCommand{
			Database:            db,
			BotRoles:            roles,
			SubmitMapsChannelID: submitMapsChannelID,
		},
		&command.UpdateCommand{
			Database: db,
			BotRoles: roles,
		},
		&command.ApproveCommand{
			Database: db,
			BotRoles: roles,
		},
		&command.DeclineCommand{
			Database: db,
			BotRoles: roles,
		},
	}
}

func SetupBot(s *discordgo.Session, db *models.Database, roles *roles.BotRoles) func() {
	submitChannel, err := channel.CreateOrGetSubmitMapChannel(s, roles)
	if err != nil {
		log.Err(err).Msg("cannot create submit channel")
		return func() {}
	}

	cmds := createCommands(db, roles, submitChannel.GetID())

	deferFunc := func() {
		for _, cmd := range cmds {
			cmd.ApplicationCommandDelete(s)
		}
	}

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		commandName := i.ApplicationCommandData().Name
		log := log.With().Str("command-name", commandName).Logger()

		handler := helpers.First(
			cmds,
			func(val command.CommandInterface) bool {
				return val.GetName() == commandName
			},
		)
		if handler == nil {
			log.Error().Msgf("handler not found for given command name")
			return
		}

		err := (*handler).Handle(s, i)
		if err != nil {
			issueID := uuid.NewString()

			_, _ = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: fmt.Sprintf("We encountered some issues during this command invocation. Please report this to an administrator. Issue ID %s", issueID),
				Flags:   discordgo.MessageFlagsEphemeral,
			})

			log.Error().Err(err).Str("issueID", issueID).Msg("cannot handle command invocation")

			return
		}
	})

	allApplication, err := s.ApplicationCommands(config.CONFIG.AppID, config.CONFIG.GuildID)
	if err != nil {
		log.Err(err).Msg("cannot retrieve all application commands")
		return deferFunc
	}

	for _, cmd := range cmds {
		e := helpers.GetFromArr(allApplication, func(e *discordgo.ApplicationCommand) bool {
			return e.Name == cmd.GetName()
		})

		if e != nil {
			log.Debug().Msgf("command '%s' already exists", cmd.GetName())
			continue
		}

		log.Debug().Msgf("command '%s' not exists", cmd.GetName())

		err := cmd.ApplicationCommandCreate(s)
		if err != nil {
			return deferFunc
		}
	}

	return deferFunc
}