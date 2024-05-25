package command

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/command/middleware"
	v2command "github.com/dixtel/dicord-bot-kog/command/v2"
)

var _ v2command.Command = (*ModCommand)(nil)

type ModCommand struct{}

func (ModCommand) Before() []middleware.CommandMiddleware {
	return []middleware.CommandMiddleware{
		middleware.CreateUser,
	}
}

func (ModCommand) Handle(
	ctx context.Context,
	opt *v2command.UserChoice,
	r *v2command.Responder,
) error {
	// r.Message().WaitForResponse()

	r.Modal().Form()

	return nil
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
