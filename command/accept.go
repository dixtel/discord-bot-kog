package command

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/channel"
	"github.com/dixtel/dicord-bot-kog/command/middleware"
	v2command "github.com/dixtel/dicord-bot-kog/command/v2"
	"github.com/dixtel/dicord-bot-kog/helpers"
	"github.com/dixtel/dicord-bot-kog/models"
	"github.com/dixtel/dicord-bot-kog/roles"
	"github.com/dixtel/dicord-bot-kog/twmap"
	"gorm.io/gorm"
)

var _ v2command.Command = (*AcceptMapCommand)(nil)

type AcceptMapCommand struct{}

func (AcceptMapCommand) Before() []middleware.CommandMiddleware {
	return []middleware.CommandMiddleware{
		middleware.CreateOrGetUser,
		middleware.InsideSubmitMapsChannel,
		middleware.UserHasMapAcceptorRole,
	}
}

func (AcceptMapCommand) Handle(
	ctx context.Context,
	opt *v2command.UserChoice,
	r *v2command.Responder,
	s *discordgo.Session,
	db *models.Database,
	botRoles *roles.BotRoles,
	channelManager *channel.ChannelManager,
) error {
	r.InteractionRespond().WaitForResponse()

	mapCreatorID, ok := opt.FromApplicationCommand().Get(AcceptMapCommand{}.GetName(), "user").(string)
	if !ok {
		return fmt.Errorf("user is nil")
	}

	m, err := db.GetLastUploadedMap(mapCreatorID)

	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		r.InteractionRespond().Content(
			"User %s doesn't have any uploaded map waiting to be accepted",
			helpers.MentionUser(mapCreatorID),
		)
		return nil
	} else if err != nil {
		return fmt.Errorf("cannot get last uploaded map: %w", err)
	}

	if m.Status != models.MapStatus_WaitingToAccept {
		r.InteractionRespond().Content(
			"User %s doesn't have any uploaded map waiting to be accepted or rejected. The status for last map is %q",
			helpers.MentionUser(mapCreatorID),
			m.Status,
		)
		return nil
	}

	testingMapChannel, err := channelManager.CreateTestingMapChannel(m.FileName, mapCreatorID)
	if err != nil {
		return fmt.Errorf("cannot create testing map channel: %w", err)
	}

	_, err = db.CreateTestingChannel(testingMapChannel.GetID(), testingMapChannel.GetName())
	if err != nil {
		return fmt.Errorf("cannot create testing channel record: %w", err)
	}

	err = db.AcceptMap(m.ID, mapCreatorID, testingMapChannel.GetID())
	if err != nil {
		_, _ = s.ChannelDelete(testingMapChannel.GetID())
		_ = db.DeleteTestingChannel(testingMapChannel.GetID())
		return fmt.Errorf("cannot mark map as accepted: %w", err)
	}

	r.MessageToChannelWithFiles(
		testingMapChannel.GetID(),
		[]*discordgo.File{

			{
				Name:        strings.Replace(m.FileName, ".map", ".png", 1),
				ContentType: "image/png",
				Reader:      bytes.NewReader(m.Screenshot),
			},
			{
				Name:        m.FileName,
				ContentType: "text/plain",
				Reader:      bytes.NewReader(m.File),
			},
		},
		"New map %s from %s!\n"+
			"%s and %s can now /vote for the map and discuss about details with the author.\n" +
			"Map author can update the map with command /update",
		twmap.RemoveMapFileExtension(m.FileName),
		helpers.MentionUser(mapCreatorID),
		helpers.MentionRole(botRoles.MapAcceptor.ID),
		helpers.MentionRole(botRoles.MapTester.ID),
	)

	r.InteractionRespond().Content(
		"Map %s from %s was accepted to the next stage - testing 🎉",
		twmap.RemoveMapFileExtension(m.FileName),
		helpers.MentionUser(mapCreatorID),
	)

	return nil
}

func (AcceptMapCommand) GetName() string {
	return "accept-map"
}

func (AcceptMapCommand) GetApplicationCommandBlueprint() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        AcceptMapCommand{}.GetName(),
		Description: AcceptMapCommand{}.GetName(),
		Options: []*discordgo.ApplicationCommandOption{
			{
		
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "Map Creator",
				Required:    true,
			},
		},
	}
}
