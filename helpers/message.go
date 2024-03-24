package helpers

import (
	"io"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

type SendMessageType string

var (
	SendMessageTypeAll        SendMessageType = "to_all"
	SendMessageTypeOriginator SendMessageType = "to_originator"
)

func SendResponse(
	msgType SendMessageType,
	msg string,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) {
	if msgType == SendMessageTypeOriginator {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: msg,
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			log.Error().Err(err).Msg("cannot send message to originator")
			return
		}
	} else if msgType == SendMessageTypeAll{
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: msg,
			},
		})
		if err != nil {
			log.Error().Err(err).Msg("cannot send message to originator")
			return
		}

	} else {
		log.Error().Msgf("wrong send message type %q", msgType)
	}
}

func SendResponseWithImage(
	msgType SendMessageType,
	msg string,
	imgName string,
	imgReader io.Reader,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) {
	if msgType == SendMessageTypeOriginator {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: msg,
				Flags:   discordgo.MessageFlagsEphemeral,
				Files: []*discordgo.File{
					{
						Name:        imgName,
						ContentType: "image/png",
						Reader:      imgReader,
					},
				},
			},
		})
		if err != nil {
			log.Error().Err(err).Msg("cannot send message to originator")
			return
		}
	} else if msgType == SendMessageTypeAll{
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: msg,
				Files: []*discordgo.File{
					{
						Name:        imgName,
						ContentType: "image/png",
						Reader:      imgReader,
					},
				},
			},
		})
		if err != nil {
			log.Error().Err(err).Msg("cannot send message to originator")
			return
		}

	} else {
		log.Error().Msgf("wrong send message type %q", msgType)
	}
}