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

var _ v2command.Command = (*UploadMapCommand)(nil)

type UploadMapCommand struct{}

func (UploadMapCommand) Before() []middleware.CommandMiddleware {
	return []middleware.CommandMiddleware{
		middleware.CreateOrGetUser,
		middleware.InsideSubmitMapsChannel,
		middleware.LastUserMapIsAccepted,
		middleware.IsNotBannedFromSubmission,
	}
}

func (UploadMapCommand) Handle(
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

	mapFileAttachment := opt.FromApplicationCommand().GetAttachment("file")
	if mapFileAttachment == nil {
		return fmt.Errorf("cannot get attachment")
	}

	if !twmap.IsValidMapFileName(mapFileAttachment.Filename) {
		r.InteractionRespond().PrivateMessage("Incorrect map filename.")
		return nil
	}

	mapExists, err := db.MapExists(mapFileAttachment.Filename)
	if err != nil {
		return fmt.Errorf("cannot check if map already exists: %w", err)
	}

	if mapExists {
		r.InteractionRespond().PrivateMessage("This map name is already taken. Please upload again with a different name")
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

	_, err = db.CreateMap(mapFileAttachment.Filename, invocator.User.ID, mapSource, screenshotSource)
	if err != nil {
		return fmt.Errorf("cannot create map %w", err)
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
		"New map %s from %s!",
		strings.Replace(mapFileAttachment.Filename, ".map", "", 1),
		helpers.MentionUser(invocator.User.ID))

	r.InteractionRespond().PrivateMessage("Map was uploaded")

	return nil
}

func (UploadMapCommand) GetName() string {
	return "upload-map"
}

func (UploadMapCommand) GetApplicationCommandBlueprint() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        UploadMapCommand{}.GetName(),
		Description: UploadMapCommand{}.GetName(),
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
