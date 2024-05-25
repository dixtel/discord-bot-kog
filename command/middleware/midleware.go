package middleware

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/models"
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
) (context.Context, error)

