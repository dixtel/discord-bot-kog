package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/channel"
	"github.com/dixtel/dicord-bot-kog/command"
	v2command "github.com/dixtel/dicord-bot-kog/command/v2"
	"github.com/dixtel/dicord-bot-kog/models"
	"github.com/dixtel/dicord-bot-kog/roles"
)

var COMMANDS = []v2command.Command{
	command.ModCommand{},
	command.SubmitMapManageCommand{},
	command.VoteForMapCommand{},
	command.UploadMapCommand{},
	command.UpdateMapCommand{},
	command.SubmitMapManageCommand{},
}

func SetupBot(
	s *discordgo.Session,
	db *models.Database,
	roles *roles.BotRoles,
	channelManager *channel.ChannelManager,
) (func(), error) {
	manager, err := v2command.NewCommandManager(s, db, roles, channelManager)
	if err != nil {
		return nil, fmt.Errorf("cannot create command manager: %w", err)
	}

	err = manager.AddCommands(COMMANDS...)
	if err != nil {
		return nil, fmt.Errorf("cannot add commands: %w", err)
	}

	manager.Start(s)

	return func() {manager.Stop(s)},nil
}
