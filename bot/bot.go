package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/command"
	"github.com/dixtel/dicord-bot-kog/helpers"
	"github.com/dixtel/dicord-bot-kog/models"
	"github.com/dixtel/dicord-bot-kog/roles"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func createCommands(db *models.Database, roles *roles.BotRoles) []command.CommandInterface {
	return []command.CommandInterface{
		&command.UploadCommand{
			Database: db,
			BotRoles: roles,
		},
		&command.AcceptCommand{
			Database: db,
			BotRoles: roles,
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

func SetupCommands(s *discordgo.Session, db *models.Database, roles *roles.BotRoles) func() {
	cmds := createCommands(db, roles)

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
			helpers.SendResponse(
				helpers.SendMessageTypeOriginator,
				fmt.Sprintf("We encountered some issues during this command invocation. Please report this to an administrator. Issue ID %s", issueID),
				s, i,
			)
			log.Error().Err(err).Str("issueID", issueID).Msg("cannot handle command invocation")

			return
		}
	})

	for _, cmd := range cmds {
		cmd.ApplicationCommandCreate(s)
	}

	return deferFunc
}
