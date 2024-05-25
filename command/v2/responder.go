package v2command

import (
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

// https://discord.com/developers/docs/interactions/receiving-and-responding#responding-to-an-interaction
type Responder struct {
	i *discordgo.InteractionCreate
	s *discordgo.Session
}

func NewResponder(i *discordgo.InteractionCreate, s *discordgo.Session) *Responder {
	return &Responder{i, s}
}

func (r *Responder) Message() *responderMessage {
	return &responderMessage{r.i, r.s}
}

func (r *Responder) Modal() *responderModal {
	return &responderModal{r.i, r.s}
}

type responderMessage struct {
	i *discordgo.InteractionCreate
	s *discordgo.Session
}

func (r *responderMessage) WaitForResponse() {
	err := r.s.InteractionRespond(r.i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: nil,
	})
	if err != nil {
		log.Error().Err(err).Msg("cannot wait for response")
	}
}

func (r *responderMessage) Content(content string) {
	_, err := r.s.FollowupMessageCreate(r.i.Interaction, true, &discordgo.WebhookParams{
		Content: content,
	})
	if err != nil {
		log.Error().Err(err).Msg("cannot respond with message to interaction")
	}
}

type responderModal struct {
	i *discordgo.InteractionCreate
	s *discordgo.Session
}

func (r *responderModal) Form() {
	err := r.s.InteractionRespond(r.i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: "test",
			Title:    "Modals survey",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "opinion",
							Label:       "What is your opinion on them?",
							Style:       discordgo.TextInputShort,
							Placeholder: "Don't be shy, share your opinion with us",
							Required:    true,
							MaxLength:   300,
							MinLength:   10,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:  "suggestions",
							Label:     "What would you suggest to improve them?",
							Style:     discordgo.TextInputParagraph,
							Required:  false,
							MaxLength: 2000,
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.Error().Err(err).Msg("cannot respond with message to interaction")
	}
}
