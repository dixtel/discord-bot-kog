package middleware

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/channel"
	"github.com/dixtel/dicord-bot-kog/models"
	"github.com/dixtel/dicord-bot-kog/roles"
)

type ErrorWithResponseToUser struct {
	MessageToUser string
}

func (e *ErrorWithResponseToUser) Error() string {
	return e.MessageToUser
}

type CommandMiddleware = func(
	context.Context,
	*discordgo.Interaction,
	*models.Database,
	*roles.BotRoles,
	*channel.ChannelManager,
) (context.Context, error)

