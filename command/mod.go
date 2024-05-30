package command

import (
	"context"
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/channel"
	"github.com/dixtel/dicord-bot-kog/command/middleware"
	v2command "github.com/dixtel/dicord-bot-kog/command/v2"
	"github.com/dixtel/dicord-bot-kog/config"
	"github.com/dixtel/dicord-bot-kog/helpers"
	"github.com/dixtel/dicord-bot-kog/models"
	"github.com/dixtel/dicord-bot-kog/roles"
	"gorm.io/gorm"
)

var _ v2command.Command = (*ModCommand)(nil)

type ModCommand struct{}

func (ModCommand) Before() []middleware.CommandMiddleware {
	return []middleware.CommandMiddleware{
		middleware.CreateOrGetUser,
		middleware.UserIsTesterOrAcceptor,
	}
}

func (ModCommand) Handle(
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

	if opt.FromApplicationCommand().Get("submission", "ban") != nil {
		userID, ok := opt.FromApplicationCommand().Get("submission", "ban", "user").(string)
		if !ok {
			return fmt.Errorf("user is nil")
		}

		reason, ok := opt.FromApplicationCommand().Get("submission", "ban", "reason").(string)
		if !ok {
			return fmt.Errorf("reason is nil")
		}

		bannedUser, err := db.GetBannedUserFromSubmission(userID)

		if err == nil {
			r.InteractionRespond().PublicMessage(
				"User %s is already banned by %s, reason: %q",
				helpers.MentionUser(userID),
				helpers.MentionUser(bannedUser.ByUserID),
				bannedUser.Reason,
			)

			return nil
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			userToBan, err := s.GuildMember(config.CONFIG.GuildID, userID)
			if err != nil {
				return fmt.Errorf("cannot get discord user: %w", err)
			}

			if botRoles.HasMapAcceptorRole(userToBan) || botRoles.HasMapTesterRole(userToBan) {
				r.InteractionRespond().PublicMessage("Cannot ban %s, he has acceptor or tester role", helpers.MentionUser(userID))
				return nil
			}

			err = db.CreateBannedUserFromSubmission(userID, reason, invocator.User.ID)
			if err != nil {
				return fmt.Errorf("cannot create banned user: %w", err)
			}

			r.InteractionRespond().PublicMessage("User %s is now banned from map submission", userID)
			return nil
		}

		return fmt.Errorf("cannot get banned user: %w", err)
	} else if opt.FromApplicationCommand().Get("submission", "unban") != nil {
		userID, ok := opt.FromApplicationCommand().Get("submission", "ban", "user").(string)
		if !ok {
			return fmt.Errorf("user is nil")
		}

		r.InteractionRespond().PublicMessage("User %s was unbanned", helpers.MentionUser(userID))

		return nil
	}

	return fmt.Errorf("cannot handle mod command")
}

func (ModCommand) GetName() string {
	return "mod"
}

func (ModCommand) GetApplicationCommandBlueprint() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        ModCommand{}.GetName(),
		Description: ModCommand{}.GetName(),
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "submission",
				Description: "submission",
				Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "ban",
						Description: "Ban user from submission",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Options: []*discordgo.ApplicationCommandOption{
							{

								Type:        discordgo.ApplicationCommandOptionUser,
								Name:        "user",
								Description: "User",
								Required:    true,
							},
							{
								Type:        discordgo.ApplicationCommandOptionString,
								Name:        "reason",
								Description: "Reason of ban",
								Required:    true,
								MinLength:   helpers.ToPtr(10),
								MaxLength:   1000,
							},
						},
					},
					{
						Name:        "unban",
						Description: "Unban user from submission",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Options: []*discordgo.ApplicationCommandOption{
							{

								Type:        discordgo.ApplicationCommandOptionUser,
								Name:        "user",
								Description: "User",
								Required:    true,
							},
						},
					},
				},
			},
		},
	}
}
