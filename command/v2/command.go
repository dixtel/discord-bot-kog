package v2command

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/channel"
	"github.com/dixtel/dicord-bot-kog/command/middleware"
	"github.com/dixtel/dicord-bot-kog/helpers"
	"github.com/dixtel/dicord-bot-kog/models"
	"github.com/dixtel/dicord-bot-kog/roles"
)

type UserChoice struct {
	i *discordgo.InteractionCreate
}

func NewUserOption(i *discordgo.InteractionCreate) *UserChoice {
	return &UserChoice{i}
}

func (up *UserChoice) FromApplicationCommand() *userChoiceFromApplicationCommand {
	if up.i.Type != discordgo.InteractionApplicationCommand {
		return &userChoiceFromApplicationCommand{d: &discordgo.ApplicationCommandInteractionData{}}
	}

	return &userChoiceFromApplicationCommand{
		helpers.ToPtr(up.i.ApplicationCommandData()),
	}
}

type userChoiceFromApplicationCommand struct {
	d *discordgo.ApplicationCommandInteractionData
}


func getOptionValue(options []*discordgo.ApplicationCommandInteractionDataOption, path ...string)  interface{} {
	for _, o := range options {
		if o.Name != path[0] {
			continue
		}

		if len(path) == 1 {
			if o.Type == discordgo.ApplicationCommandOptionSubCommand {
				return struct{}{}
			}

			return o.Value
		}

		return getOptionValue(o.Options, path[1:]...)
	}

	return nil
}

func (u *userChoiceFromApplicationCommand) Get(path ...string) interface{} {
	return getOptionValue(u.d.Options, path...)
}

type Command interface {
	Before() []middleware.CommandMiddleware
	Handle(
		context.Context,
		*UserChoice,
		*Responder,
		*discordgo.Session,
		*models.Database,
		*roles.BotRoles,
		*channel.ChannelManager,
	) error
	GetName() string
	GetApplicationCommandBlueprint() *discordgo.ApplicationCommand
}
