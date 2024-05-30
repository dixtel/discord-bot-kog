package v2command

import (
	"context"
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/channel"
	"github.com/dixtel/dicord-bot-kog/command/middleware"
	"github.com/dixtel/dicord-bot-kog/config"
	"github.com/dixtel/dicord-bot-kog/helpers"
	"github.com/dixtel/dicord-bot-kog/models"
	"github.com/dixtel/dicord-bot-kog/roles"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type CommandManager struct {
	cmds                    []Command
	rawCmds                 map[string]*discordgo.ApplicationCommand
	s                       *discordgo.Session
	db                      *models.Database
	botRoles                *roles.BotRoles
	channelManager          *channel.ChannelManager
	allRawCommandsOnStartup []*discordgo.ApplicationCommand
}

func NewCommandManager(
	s *discordgo.Session,
	db *models.Database,
	botRoles *roles.BotRoles,
	channelManager *channel.ChannelManager,
) (*CommandManager, error) {
	err := removeCommands(s)
	if err != nil {
		return nil, fmt.Errorf("remove commands: %w", err)
	}

	allRawCommands, err := s.ApplicationCommands(
		config.CONFIG.AppID,
		config.CONFIG.GuildID,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve discord applications commands: %w", err)
	}

	return &CommandManager{
		cmds:                    []Command{},
		rawCmds:                 map[string]*discordgo.ApplicationCommand{},
		s:                       s,
		db:                      db,
		botRoles:                botRoles,
		channelManager:          channelManager,
		allRawCommandsOnStartup: allRawCommands,
	}, nil
}

func (m *CommandManager) AddCommands(cmds ...Command) error {
	for _, cmd := range cmds {
		if err := m.createApplicationCommand(cmd); err != nil {
			return fmt.Errorf("cannot add command: %w", err)
		}
	}

	m.cmds = append(m.cmds, cmds...)

	return nil
}

func (m *CommandManager) Start(s *discordgo.Session) {
	modalSubmitHandler := NewModalSubmitHandler()
	modalSubmitHandler.Start(s)

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Interaction.Type != discordgo.InteractionApplicationCommand {
			return
		}

		commandName := i.ApplicationCommandData().Name
		log := log.With().Str("command-name", commandName).Logger()

		command := helpers.First(
			m.cmds,
			func(val Command) bool {
				return val.GetName() == commandName
			},
		)
		if command == nil {
			log.Error().Msgf("command %q not exists", commandName)
			return
		}

		commandCtx := context.Background()
		responder := NewResponder(i, s, modalSubmitHandler)

		responder.InteractionRespond().WaitForResponse()

		for _, middlewareHandler := range command.Before() {
			ctx, err := middlewareHandler(
				commandCtx,
				i.Interaction,
				m.db,
				m.botRoles,
				m.channelManager,
			)
			commandCtx = ctx

			if err != nil {
				var errorWithResponseToUser *middleware.ErrorWithResponseToUser
				if errors.As(err, &errorWithResponseToUser) {
					responder.InteractionRespond().PrivateMessage("⛔ %s", errorWithResponseToUser.MessageToUser)
					return
				}

				reportErrorToUser(
					responder,
					fmt.Errorf("command middleware returns an error: %w", err),
					log.Error().Str("middleware-func", helpers.GetFunctionName(middlewareHandler)),
				)
				return
			}
		}

		db, commit, rollback := m.db.TxV2()
		err := command.Handle(
			commandCtx,
			NewUserOption(i),
			responder,
			s,
			db,
			m.botRoles,
			m.channelManager,
		)
		if err != nil {
			rollback()
			reportErrorToUser(
				responder,
				fmt.Errorf("command handler returns an error: %w", err),
				log.Error(),
			)
			return
		}

		commit()
	})
}

func (m *CommandManager) Stop(s *discordgo.Session) {
	_ = removeCommands(s)
}

func (m *CommandManager) createApplicationCommand(cmd Command) error {
	rawCommand := helpers.First(
		m.allRawCommandsOnStartup,
		func(c *discordgo.ApplicationCommand) bool {
			return c.Name == cmd.GetName()
		},
	)

	if rawCommand == nil {
		raw, err := m.s.ApplicationCommandCreate(
			config.CONFIG.AppID,
			config.CONFIG.GuildID,
			cmd.GetApplicationCommandBlueprint(),
		)
		if err != nil {
			return fmt.Errorf("cannot create application command %q: %w", cmd.GetName(), err)
		}

		rawCommand = raw

		log.Info().Msgf("command %s was created", rawCommand.Name)
	} else {
		log.Info().Msgf("command %s already exists", rawCommand.Name)
	}

	m.rawCmds[cmd.GetName()] = rawCommand

	return nil
}

func reportErrorToUser(r *Responder, err error, event *zerolog.Event) {
	issueID := uuid.NewString()
	r.InteractionRespond().
		PrivateMessage(fmt.Sprintf(
			"We encountered some issues during command invocation.\nPlease report this to an administrator.\nIssue ID %s",
			issueID,
		),
		)
	event.Err(err).Str("issue-id", issueID).Msg("cannot handle command")
}


func removeCommands(s *discordgo.Session) error {
	if config.CONFIG.Env == "dev" {
		return nil
	}

	allRawCommands, err := s.ApplicationCommands(
		config.CONFIG.AppID,
		config.CONFIG.GuildID,
	)
	if err != nil {
		return fmt.Errorf("cannot retrieve discord applications commands: %w", err)
	}

	for _, rawCmd := range allRawCommands {
		err := s.ApplicationCommandDelete(
			config.CONFIG.AppID,
			config.CONFIG.GuildID,
			rawCmd.ID,
		)
		if err != nil {
			return fmt.Errorf("cannot remove command: %w", err)
		}

		log.Info().Msgf("command %s was removed", rawCmd.Name)
	}

	return nil
}