package middleware

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/channel"
	"github.com/dixtel/dicord-bot-kog/models"
	"github.com/dixtel/dicord-bot-kog/roles"
)

var IsNotBannedFromSubmission CommandMiddleware = func(
	ctx context.Context,
	i *discordgo.Interaction,
	db *models.Database,
	botRoles *roles.BotRoles,
	_ *channel.ChannelManager,
) (context.Context, error) {
	bannedUser, err := db.GetBannedUserFromSubmission(i.Member.User.ID)
	if err == nil {
		return nil, &ErrorWithResponseToUser{
			MessageToUser: fmt.Sprintf("You are banned from map submission, reason: %s", bannedUser.Reason),
		}
	}

	return ctx, nil
}

type CreateUserContext struct {
	User *models.User
}

var CreateOrGetUser CommandMiddleware = func(
	ctx context.Context,
	i *discordgo.Interaction,
	db *models.Database,
	botRoles *roles.BotRoles,
	_ *channel.ChannelManager,
) (context.Context, error) {
	user, err := db.CreateOrGetUser(i.Member.User.Username, i.Member.User.ID)
	if err != nil {
		return nil, fmt.Errorf("cannot create or get user: %w", err)
	}

	return context.WithValue(ctx, CreateUserContext{}, CreateUserContext{user}), nil
}
