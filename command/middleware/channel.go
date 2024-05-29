package middleware

import (
	"context"
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/channel"
	"github.com/dixtel/dicord-bot-kog/helpers"
	"github.com/dixtel/dicord-bot-kog/models"
	"github.com/dixtel/dicord-bot-kog/roles"
	"gorm.io/gorm"
)

var InsideSubmitMapsChannel CommandMiddleware = func(
	ctx context.Context,
	i *discordgo.Interaction,
	db *models.Database,
	botRoles *roles.BotRoles,
	channelManager *channel.ChannelManager,
) (context.Context, error) {
	if i.ChannelID != channelManager.GetSubmitMapChannelID() {
		return nil, &ErrorWithResponseToUser{
			MessageToUser: fmt.Sprintf(
				"You need to be inside %s to use this command",
				helpers.MentionChannel(channelManager.GetSubmitMapChannelID()),
			),
		}
	}

	return ctx, nil
}

var InsideTestingMapChannel CommandMiddleware = func(
	ctx context.Context,
	i *discordgo.Interaction,
	db *models.Database,
	botRoles *roles.BotRoles,
	channelManager *channel.ChannelManager,
) (context.Context, error) {
	isTestingChannel, err := db.IsTestingChannel(i.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("cannot check if current channel is a testing channel: %w", err)
	}

	if isTestingChannel {
		return ctx, nil
	}

	return nil, &ErrorWithResponseToUser{
		MessageToUser: "This is not a testing channel",
	}
}

type TestingChannelDataContext struct {
	Map                *models.Map
	TestingChannelData *models.TestingChannelData
}

var TestingChannelData CommandMiddleware = func(
	ctx context.Context,
	i *discordgo.Interaction,
	db *models.Database,
	botRoles *roles.BotRoles,
	channelManager *channel.ChannelManager,
) (context.Context, error) {
	m, err := db.GetLastUploadedMapByChannelID(i.ChannelID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, &ErrorWithResponseToUser{
			MessageToUser: "No maps found",
		}
	}

	if err != nil {
		return nil, fmt.Errorf("cannot get map: %w", err)
	}

	data, err := db.GetTestingChannelData(m.ID)
	if err != nil {
		return nil, fmt.Errorf("cannot get testing channel data %w", err)
	}

	return context.WithValue(ctx, TestingChannelDataContext{}, TestingChannelDataContext{
		Map:                m,
		TestingChannelData: data,
	}), nil
}
