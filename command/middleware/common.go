package middleware

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/channel"
	"github.com/dixtel/dicord-bot-kog/models"
	"github.com/dixtel/dicord-bot-kog/roles"
)

type CommonUtilsMiddlewareContext struct {
	DB *models.Database
	Roles *roles.BotRoles
	ChannelManager *channel.ChannelManager
	I *discordgo.Interaction
}

var CommonUtilsMiddleware CommandMiddleware = func(
	ctx context.Context,
	i *discordgo.Interaction,
	db *models.Database,
	botRoles *roles.BotRoles,
	channelManager *channel.ChannelManager,
) (context.Context, error) {
	return context.WithValue(ctx, CommonUtilsMiddlewareContext{}, CommonUtilsMiddlewareContext{
		DB:             db,
		Roles:          botRoles,
		ChannelManager: channelManager,
		I:              i,
	}), nil
}
