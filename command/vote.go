package command

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/channel"
	"github.com/dixtel/dicord-bot-kog/command/middleware"
	v2command "github.com/dixtel/dicord-bot-kog/command/v2"
	"github.com/dixtel/dicord-bot-kog/config"
	"github.com/dixtel/dicord-bot-kog/helpers"
	"github.com/dixtel/dicord-bot-kog/models"
	"github.com/dixtel/dicord-bot-kog/roles"
	"github.com/rs/zerolog/log"
)

var _ v2command.Command = (*VoteForMapCommand)(nil)

type VoteForMapCommand struct{}

func (VoteForMapCommand) Before() []middleware.CommandMiddleware {
	return []middleware.CommandMiddleware{
		middleware.CreateOrGetUser,
		middleware.UserIsTesterOrAcceptor,
		middleware.InsideTestingMapChannel,
		middleware.TestingChannelData,
	}
}

func (VoteForMapCommand) Handle(
	ctx context.Context,
	opt *v2command.UserChoice,
	r *v2command.Responder,
	s *discordgo.Session,
	db *models.Database,
	botRoles *roles.BotRoles,
	channelManager *channel.ChannelManager,
) error {
	testingChannelDataContext, ok := ctx.Value(middleware.TestingChannelDataContext{}).(middleware.TestingChannelDataContext)
	if !ok {
		return fmt.Errorf("TestingChannelDataContext is not set")
	}

	if testingChannelDataContext.Map.Status != models.MapStatus_Accepted {
		r.InteractionRespond().PublicMessage("Voting for this map is disabled")
		return nil
	}

	if opt.FromApplicationCommand().Get("vote", "up") != nil {
		return VoteForMapCommand{}.voteUp(ctx, db, r, channelManager, s)
	} else if opt.FromApplicationCommand().Get("vote", "down") != nil {
		reason, ok := opt.FromApplicationCommand().Get("vote", "down", "reason").(string)
		if !ok {
			return fmt.Errorf("reason is nil")
		}

		return VoteForMapCommand{}.voteDown(ctx, db, reason, r)
	}

	return fmt.Errorf("cannot handle vote command")
}

func (VoteForMapCommand) voteUp(
	ctx context.Context,
	db *models.Database,
	r *v2command.Responder,
	channelManager *channel.ChannelManager,
	s *discordgo.Session,
) error {
	testingChannelDataContext, ok := ctx.Value(middleware.TestingChannelDataContext{}).(middleware.TestingChannelDataContext)
	if !ok {
		return fmt.Errorf("TestingChannelDataContext is not set")
	}

	createUserContext, ok := ctx.Value(middleware.CreateUserContext{}).(middleware.CreateUserContext)
	if !ok {
		return fmt.Errorf("CreateUserContext is not set")
	}

	delete(testingChannelDataContext.TestingChannelData.DeclinedBy, createUserContext.User.ID)
	testingChannelDataContext.TestingChannelData.ApprovedBy[createUserContext.User.ID] = struct{}{}

	err := db.UpdateTestingChannelData(testingChannelDataContext.Map.ID, testingChannelDataContext.TestingChannelData)
	if err != nil {
		return fmt.Errorf("cannot update testing channel data %w", err)
	}

	r.InteractionRespond().PublicMessage("Up vote from %s 👍", helpers.MentionUser(createUserContext.User.ID))

	return VoteForMapCommand{}.tryApproveMap(ctx, db, &testingChannelDataContext, r, channelManager, s)
}

func (VoteForMapCommand) voteDown(
	ctx context.Context,
	db *models.Database,
	reason string,
	r *v2command.Responder,
) error {
	testingChannelDataContext, ok := ctx.Value(middleware.TestingChannelDataContext{}).(middleware.TestingChannelDataContext)
	if !ok {
		return fmt.Errorf("TestingChannelDataContext is not set")
	}

	createUserContext, ok := ctx.Value(middleware.CreateUserContext{}).(middleware.CreateUserContext)
	if !ok {
		return fmt.Errorf("CreateUserContext is not set")
	}

	delete(testingChannelDataContext.TestingChannelData.ApprovedBy, createUserContext.User.ID)
	testingChannelDataContext.TestingChannelData.DeclinedBy[createUserContext.User.ID] = struct{}{}

	err := db.UpdateTestingChannelData(testingChannelDataContext.Map.ID, testingChannelDataContext.TestingChannelData)
	if err != nil {
		return fmt.Errorf("cannot update testing channel data %w", err)
	}

	r.InteractionRespond().PublicMessage("Down vote from %s 👎, reason: %s", helpers.MentionUser(createUserContext.User.ID), reason)

	approvedCounter := len(testingChannelDataContext.TestingChannelData.ApprovedBy)
	declinedCounter := len(testingChannelDataContext.TestingChannelData.DeclinedBy)

	approvalsRemaining := (config.CONFIG.MinimumMapApprovalsNumber + 1) - (approvedCounter - declinedCounter)
	r.InteractionRespond().PublicMessage(
		"Needs %v approvals to accept this map. (approvals %v / declines %v)",
		approvalsRemaining, approvedCounter, declinedCounter,
	)

	return nil
}

func (VoteForMapCommand) tryApproveMap(
	ctx context.Context,
	db *models.Database,
	data *middleware.TestingChannelDataContext,
	r *v2command.Responder,
	channelManager *channel.ChannelManager,
	s *discordgo.Session,
) error {
	approvedCounter := len(data.TestingChannelData.ApprovedBy)
	declinedCounter := len(data.TestingChannelData.DeclinedBy)

	isApproved := approvedCounter-declinedCounter > config.CONFIG.MinimumMapApprovalsNumber
	if !isApproved {
		approvalsRemaining := (config.CONFIG.MinimumMapApprovalsNumber + 1) - (approvedCounter - declinedCounter)
		r.InteractionRespond().PublicMessage(
			"Needs %v approvals to accept this map. (approvals %v / declines %v)",
			approvalsRemaining, approvedCounter, declinedCounter,
		)

		return nil
	}

	err := db.ApproveMap(data.Map.ID)
	if err != nil {
		return fmt.Errorf("cannot mark map as approved %w", err)
	}

	r.MessageToChannelWithFiles(
		channelManager.GetSubmitMapChannelID(),
		[]*discordgo.File{
			{
				Name:        strings.Replace(string(data.Map.FileName), ".map", ".png", 1),
				ContentType: "image.png",
				Reader:      bytes.NewReader(data.Map.Screenshot),
			},
		},
		"Map from %s was successfully approved by testers 🕺",
		helpers.MentionUser(data.Map.MapperID),
	)

	r.InteractionRespond().PublicMessage(
		"Map was successfully approved by testers! Congratulation %s 🎉\n"+
			"This channel will be removed soon automatically",
		helpers.MentionUser(data.Map.MapperID),
	)

	go func() {
		time.Sleep(time.Hour * 24)

		_, err := s.ChannelDelete(data.Map.TestingChannel.ChannelID)
		if err != nil {
			log.
				Error().
				Err(err).
				Msgf("cannot remove channel %q after approval", data.Map.TestingChannel.ChannelID)
		}
	}()

	return nil
}

func (VoteForMapCommand) GetName() string {
	return "vote"
}

func (VoteForMapCommand) GetApplicationCommandBlueprint() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        VoteForMapCommand{}.GetName(),
		Description: VoteForMapCommand{}.GetName(),
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "vote",
				Description: "Vote Up/Down",
				Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "up",
						Description: "Vote up",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
					},
					{
						Name:        "down",
						Description: "Vote down",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Options: []*discordgo.ApplicationCommandOption{
							{
								Type:        discordgo.ApplicationCommandOptionString,
								Name:        "reason",
								Description: "Reason",
								Required:    true,
								MinLength:   helpers.ToPtr(10),
								MaxLength:   1000,
							},
						},
					},
				},
			},
		},
	}
}
