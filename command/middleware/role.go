package middleware

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/channel"
	"github.com/dixtel/dicord-bot-kog/helpers"
	"github.com/dixtel/dicord-bot-kog/models"
	"github.com/dixtel/dicord-bot-kog/roles"
)

var UserIsTesterOrAcceptor CommandMiddleware = func(
	ctx context.Context,
	i *discordgo.Interaction,
	db *models.Database,
	botRoles *roles.BotRoles,
	_ *channel.ChannelManager,
) (context.Context, error) {
	if !(botRoles.HasMapAcceptorRole(i.Member) || botRoles.HasMapTesterRole(i.Member)) {
		return nil, &ErrorWithResponseToUser{
			MessageToUser: "You don't have acceptor or tester role to use this command",
		}
	}

	return ctx, nil
}

var UserHasMapAcceptorRole CommandMiddleware = func(
	ctx context.Context,
	i *discordgo.Interaction,
	db *models.Database,
	botRoles *roles.BotRoles,
	_ *channel.ChannelManager,
) (context.Context, error) {
	if !botRoles.HasMapAcceptorRole(i.Member) {
		return nil, &ErrorWithResponseToUser{
			MessageToUser: fmt.Sprintf(
				"You must have %s role to accept maps",
				helpers.MentionRole(botRoles.MapAcceptor.ID),
			),
		}
	}

	return ctx, nil
}
