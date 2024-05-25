package middleware

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/models"
)


var UserIsTester CommandMiddleware = func(
	ctx context.Context,
	i *discordgo.Interaction,
	db *models.Database,) (context.Context, error) {
	return nil, nil
}
