package bot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/command"
	"github.com/dixtel/dicord-bot-kog/helpers"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)


func createCommands(db *gorm.DB) []command.CommandInterface {
	return []command.CommandInterface{
		&command.UploadCommand{},
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

func SetupCommands(s *discordgo.Session, db *gorm.DB) func() {
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
			log.Error().Msgf("handler not found for command name")
			return
		}

		err := (*handler).Handle(s, i)
		if err != nil {
			log.Error().Err(err).Msgf("cannot handle command invocation")
			return
		}
	})

	for _, cmd := range cmds {
		cmd.ApplicationCommandCreate(s)
	}

	return deferFunc
}
