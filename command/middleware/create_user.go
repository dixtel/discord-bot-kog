package middleware

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/models"
)

type CreateUserContext struct {
	User *models.User
}

var CreateUser CommandMiddleware = func(
	ctx context.Context,
	i *discordgo.Interaction,
	db *models.Database,
) (context.Context, error) {
	user, err := db.CreateOrGetUser(i.Member.User.Username, i.Member.User.ID)
	if err != nil {
		return nil, fmt.Errorf("cannot create or get user: %w", err)
	}

	return context.WithValue(ctx, CreateUserContext{}, CreateUserContext{user}), nil
}
