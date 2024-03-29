package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/command"
	"github.com/dixtel/dicord-bot-kog/helpers"
	"github.com/dixtel/dicord-bot-kog/models"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func createCommands(db *models.Database) []command.CommandInterface {
	return []command.CommandInterface{
		&command.UploadCommand{
			Database: db,
		},
		&command.AcceptCommand{
			Database: db,
		},
		&command.UpdateCommand{
			Database: db,
		},
		&command.ApproveCommand{
			Database: db,
		},
		&command.DeclineCommand{
			Database: db,
		},
	}

	// return []*Command{
	// 	{
	// 		Name:        "upload",
	// 		Description: "Upload a new map",
	// 		Handler: func(_ *discordgo.InteractionCreate, sendMessage func(SendMessageType, string)) error {

	// 			return nil
	// 		},
	// 	},
	// 	{
	// 		Name:        "accept",
	// 		Description: "Accept the map. This command will create a new channel for map discussion and approval",
	// 		Handler: func(_ *discordgo.InteractionCreate, sendMessage func(SendMessageType, string)) error {

	// 			return nil
	// 		},
	// 	},
	// 	{
	// 		Name:        "approve",
	// 		Description: "Approve the map",
	// 		Handler: func(_ *discordgo.InteractionCreate, sendMessage func(SendMessageType, string)) error {

	// 			return nil
	// 		},
	// 	},
	// 	{
	// 		Name:        "decline",
	// 		Description: "Decline the map",
	// 		Handler: func(_ *discordgo.InteractionCreate, sendMessage func(SendMessageType, string)) error {

	// 			return nil
	// 		},
	// 	},
	// 	{
	// 		Name:        "update",
	// 		Description: "Update the map",
	// 		Handler: func(_ *discordgo.InteractionCreate, sendMessage func(SendMessageType, string)) error {

	// 			return nil
	// 		},
	// 	},
	// }
}

func SetupCommands(s *discordgo.Session, db *models.Database) func() {
	cmds := createCommands(db)

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
