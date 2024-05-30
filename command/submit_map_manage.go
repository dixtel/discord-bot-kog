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

var _ v2command.Command = (*SubmitMapManageCommand)(nil)

type SubmitMapManageCommand struct{}

func (SubmitMapManageCommand) Before() []middleware.CommandMiddleware {
	return []middleware.CommandMiddleware{
		middleware.CreateOrGetUser,
		middleware.InsideSubmitMapsChannel,
		middleware.UserHasMapAcceptorRole,
		middleware.CommonUtilsMiddleware,
	}
}

func (SubmitMapManageCommand) Handle(
	ctx context.Context,
	opt *v2command.UserChoice,
	r *v2command.Responder,
	s *discordgo.Session,
	db *models.Database,
	botRoles *roles.BotRoles,
	channelManager *channel.ChannelManager,
) error {
	if opt.FromApplicationCommand().Get("accept") != nil {
		mapCreatorID, ok := opt.FromApplicationCommand().Get("accept", "user").(string)
		if !ok {
			return fmt.Errorf("user is nil")
		}

		return SubmitMapManageCommand{}.Accept(ctx, mapCreatorID, r, s)
	} else if opt.FromApplicationCommand().Get("reject") != nil {
		mapCreatorID, ok := opt.FromApplicationCommand().Get("reject", "user").(string)
		if !ok {
			return fmt.Errorf("user is nil")
		}

		return SubmitMapManageCommand{}.Reject(ctx, mapCreatorID, r, s)
	} 


	return fmt.Errorf("cannot handle command")
}
func (SubmitMapManageCommand) Accept(
	ctx context.Context, 
	mapCreatorID string,
	r *v2command.Responder,
	s *discordgo.Session,
) error {
	utils, ok := ctx.Value(middleware.CommonUtilsMiddlewareContext{}).(middleware.CommonUtilsMiddlewareContext)
	if !ok {
		return fmt.Errorf("CommonUtilsMiddlewareContext is not set")
	}

	m, err := utils.DB.GetLastUploadedMap(mapCreatorID)

	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		r.InteractionRespond().PublicMessage(
			"User %s doesn't have any uploaded map waiting to be accepted",
			helpers.MentionUser(mapCreatorID),
		)
		return nil
	} else if err != nil {
		return fmt.Errorf("cannot get last uploaded map: %w", err)
	}

	if m.Status != models.MapStatus_WaitingToAccept {
		r.InteractionRespond().PublicMessage(
			"User %s doesn't have any uploaded map waiting to be accepted. The status for the last map is %q",
			helpers.MentionUser(mapCreatorID),
			m.Status,
		)
		return nil
	}

	testingMapChannel, err := utils.ChannelManager.CreateTestingMapChannel(m.FileName, mapCreatorID)
	if err != nil {
		return fmt.Errorf("cannot create testing map channel: %w", err)
	}

	_, err = utils.DB.CreateTestingChannel(testingMapChannel.GetID(), testingMapChannel.GetName())
	if err != nil {
		return fmt.Errorf("cannot create testing channel record: %w", err)
	}

	err = utils.DB.AcceptMap(m.ID, mapCreatorID, testingMapChannel.GetID())
	if err != nil {
		_, _ = s.ChannelDelete(testingMapChannel.GetID())
		_ = utils.DB.DeleteTestingChannel(testingMapChannel.GetID())
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
		helpers.MentionRole(utils.Roles.MapAcceptor.ID),
		helpers.MentionRole(utils.Roles.MapTester.ID),
	)

	r.InteractionRespond().PublicMessage(
		"Map %s from %s was accepted to the next stage - testing 🎉",
		twmap.RemoveMapFileExtension(m.FileName),
		helpers.MentionUser(mapCreatorID),
	)

	return nil
}

func (SubmitMapManageCommand) Reject(ctx context.Context, 
	mapCreatorID string,
	r *v2command.Responder,
	s *discordgo.Session,
) error {
	utils, ok := ctx.Value(middleware.CommonUtilsMiddlewareContext{}).(middleware.CommonUtilsMiddlewareContext)
	if !ok {
		return fmt.Errorf("CommonUtilsMiddlewareContext is not set")
	}

	m, err := utils.DB.GetLastUploadedMap(mapCreatorID)

	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		r.InteractionRespond().PublicMessage(
			"User %s doesn't have any uploaded map waiting to be accepted",
			helpers.MentionUser(mapCreatorID),
		)
		return nil
	} else if err != nil {
		return fmt.Errorf("cannot get last uploaded map: %w", err)
	}

	if m.Status != models.MapStatus_WaitingToAccept {
		r.InteractionRespond().PublicMessage(
			"User %s doesn't have any uploaded map waiting to be rejected. The status for the last map is %q",
			helpers.MentionUser(mapCreatorID),
			m.Status,
		)
		return nil
	}

	err = utils.DB.RejectMap(m.ID, mapCreatorID, m.ID)
	if err != nil {
		return fmt.Errorf("cannot mark map as rejected: %w", err)
	}

	r.InteractionRespond().PublicMessage(
		"Map %s from %s was rejected",
		twmap.RemoveMapFileExtension(m.FileName),
		helpers.MentionUser(mapCreatorID),
	)

	return nil
}

func (SubmitMapManageCommand) GetName() string {
	return "submit-map"
}

func (SubmitMapManageCommand) GetApplicationCommandBlueprint() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        SubmitMapManageCommand{}.GetName(),
		Description: SubmitMapManageCommand{}.GetName(),
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "accept",
				Description: "Accept",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "user",
						Required:    true,
					},
				},
			},
			{
				Name:        "reject",
				Description: "Reject",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "user",
						Required:    true,
					},
				},
			},
		},
	}
}
