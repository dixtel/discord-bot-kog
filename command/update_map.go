package command

import (
	"bytes"
	"context"
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
)

var _ v2command.Command = (*UpdateMapCommand)(nil)

type UpdateMapCommand struct{}

func (UpdateMapCommand) Before() []middleware.CommandMiddleware {
	return []middleware.CommandMiddleware{
		middleware.CreateOrGetUser,
		middleware.InsideTestingMapChannel,
		middleware.LastUserMapIsAccepted,
		middleware.TestingChannelData,
	}
}

func (UpdateMapCommand) Handle(
	ctx context.Context,
	opt *v2command.UserChoice,
	r *v2command.Responder,
	s *discordgo.Session,
	db *models.Database,
	botRoles *roles.BotRoles,
	channelManager *channel.ChannelManager,
) error {
	invocator, ok := ctx.Value(middleware.CreateUserContext{}).(middleware.CreateUserContext)
	if !ok {
		return fmt.Errorf("CreateUserContext is not set")
	}

	testingChannelDataContext, ok := ctx.Value(middleware.TestingChannelDataContext{}).(middleware.TestingChannelDataContext)
	if !ok {
		return fmt.Errorf("TestingChannelDataContext is not set")
	}

	if invocator.User.ID != testingChannelDataContext.Map.MapperID {
		r.InteractionRespond().PrivateMessage("Cannot update because you are not owner of this map")
		return nil	
	}

	mapFileAttachment := opt.FromApplicationCommand().GetAttachment("file")
	if mapFileAttachment == nil {
		return fmt.Errorf("cannot get attachment")
	}

	m, err := db.GetLastUploadedMap(invocator.User.ID)
	if err != nil {
		return fmt.Errorf("cannot get last uploaded map: %w", err)
	}

	if m.Status != models.MapStatus_Accepted {
		r.InteractionRespond().PrivateMessage("Cannot update the map. Current map status is %q", m.Status)
		return nil	
	}

	if m.FileName != mapFileAttachment.Filename {
		r.InteractionRespond().PrivateMessage(
			"Map file should be named %q instead of %q", 
			m.FileName, 
			mapFileAttachment.Filename,
		)
		return nil	
	}

	mapSource, err := twmap.DownloadMapFromDiscord(mapFileAttachment.URL)
	if err != nil {
		return fmt.Errorf("cannot download map from discord %w", err)
	}

	screenshotSource, err := twmap.MakeScreenshot(mapSource)
	if err != nil {
		return fmt.Errorf("cannot make screenshot of the map: %w", err)
	}

	err = db.UpdateMap(m.ID, mapSource, screenshotSource)
	if err != nil {
		return fmt.Errorf("cannot update map %w", err)
	}

	err = db.UpdateTestingChannelData(m.ID, &models.TestingChannelData{
		ApprovedBy: map[string]struct{}{},
		DeclinedBy: map[string]struct{}{},
	})
	if err != nil {
		return fmt.Errorf("cannot update testing channel data %w", err)
	}

	r.MessageToChannelWithFiles(
		channelManager.GetSubmitMapChannelID(),
		[]*discordgo.File{
			{
				Name:        strings.Replace(mapFileAttachment.Filename, ".map", ".png", 1),
				ContentType: "image/png",
				Reader:      bytes.NewReader(screenshotSource),
			},
			{
				Name:        mapFileAttachment.Filename,
				ContentType: "text/plain",
				Reader:      bytes.NewReader(mapSource),
			},
		},
		"Map was updated by %s!\nApprovals and declines were reset",
		helpers.MentionUser(invocator.User.ID))

	r.InteractionRespond().PrivateMessage("Map was updated")

	return nil
}

func (UpdateMapCommand) GetName() string {
	return "update-map"
}

func (UpdateMapCommand) GetApplicationCommandBlueprint() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        UpdateMapCommand{}.GetName(),
		Description: UpdateMapCommand{}.GetName(),
		Options: []*discordgo.ApplicationCommandOption{
			{

				Type:        discordgo.ApplicationCommandOptionAttachment,
				Name:        "file",
				Description: "Map file",
				Required:    true,
			},
		},
	}
}
