package command

import (
	"bytes"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type Interaction struct {
	session            *discordgo.Session
	discordInteraction *discordgo.InteractionCreate
}

func FromDiscordInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) (*Interaction, error) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: "Waiting for response...",
		},
	})

	if err != nil {
		return nil, fmt.Errorf("cannot respond to interaction: %w", err)
	}

	return &Interaction{session: s, discordInteraction: i}, nil
}

type InteractionMessageType = int

const (
	InteractionMessageType_Private InteractionMessageType = 0
	InteractionMessageType_Public  InteractionMessageType = 1
)

func (i *Interaction) SendMessage(msg string, messageType InteractionMessageType) error {
	flags := discordgo.MessageFlags(0)

	if messageType == InteractionMessageType_Private {
		flags = discordgo.MessageFlagsEphemeral
	}

	_, err := i.session.FollowupMessageCreate(i.discordInteraction.Interaction, true, &discordgo.WebhookParams{
		Content: msg,
		Flags:   flags,
	})
	if err != nil {
		return fmt.Errorf("cannot send follow up message: %w", err)
	}

	return nil
}

func (i *Interaction) SendMessageWithPNGImage(
	msg string,
	messageType InteractionMessageType,
	filename string,
	src []byte,
) error {
	flags := discordgo.MessageFlags(0)

	if messageType == InteractionMessageType_Private {
		flags = discordgo.MessageFlagsEphemeral
	}

	_, err := i.session.FollowupMessageCreate(i.discordInteraction.Interaction, true, &discordgo.WebhookParams{
		Content: msg,
		Flags:   flags,
		Files: []*discordgo.File{
			{
				Name:        filename,
				ContentType: "image/png",
				Reader:      bytes.NewReader(src),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("cannot send follow up message: %w", err)
	}

	return nil
}
