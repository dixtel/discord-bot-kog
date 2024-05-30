package v2command

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/helpers"
	"github.com/rs/zerolog/log"
)

// https://discord.com/developers/docs/interactions/receiving-and-responding#responding-to-an-interaction
type Responder struct {
	i   *discordgo.InteractionCreate
	s   *discordgo.Session
	msh *ModalSubmitHandler
}

func NewResponder(
	i *discordgo.InteractionCreate,
	s *discordgo.Session,
	msh *ModalSubmitHandler,
) *Responder {
	return &Responder{i, s, msh}
}

func (r *Responder) MessageToChannel(
	channelID string,
	format string,
	args ...interface{},
) {
	_, err := r.s.ChannelMessageSend(channelID, fmt.Sprintf(format, args...))
	if err != nil {
		log.Error().Err(err).Msg("cannot send message and files to channel")
	}
}

func (r *Responder) MessageToChannelWithFiles(
	channelID string,
	files []*discordgo.File,
	format string,
	args ...interface{},
) {
	_, err := r.s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content: fmt.Sprintf(format, args...),
		Files:   files,
	})
	if err != nil {
		log.Error().Err(err).Msg("cannot send message and files to channel")
	}
}

func (r *Responder) InteractionRespond() *interactionRespond {
	return &interactionRespond{r.i, r.s}
}

func (r *Responder) ModalRespond() *modalRespond {
	return &modalRespond{r.i, r.s, r.msh}
}

type interactionRespond struct {
	i *discordgo.InteractionCreate
	s *discordgo.Session
}

func (r *interactionRespond) WaitForResponse() {
	err := r.s.InteractionRespond(r.i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: nil,
	})
	if err != nil {
		log.Error().Err(err).Msg("cannot wait for response")
	}
}

func (r *interactionRespond) PublicMessage(format string, args ...interface{}) {
	_, err := r.s.FollowupMessageCreate(r.i.Interaction, true, &discordgo.WebhookParams{
		Content: fmt.Sprintf(format, args...),
	})
	if err != nil {
		log.Error().Err(err).Msg("cannot respond with message to interaction")
	}
}

func (r *interactionRespond) PrivateMessage(format string, args ...interface{}) {
	_, err := r.s.FollowupMessageCreate(r.i.Interaction, true, &discordgo.WebhookParams{
		Content: fmt.Sprintf(format, args...),
		Flags: discordgo.MessageFlagsEphemeral,
	})
	if err != nil {
		log.Error().Err(err).Msg("cannot respond with message to interaction")
	}
}

func (r *interactionRespond) ContentWithFiles(
	files []*discordgo.File,
	format string,
	args ...interface{},
) {
	_, err := r.s.FollowupMessageCreate(r.i.Interaction, true, &discordgo.WebhookParams{
		Content: fmt.Sprintf(format, args...),
		Files:   files,
	})
	if err != nil {
		log.Error().Err(err).Msg("cannot respond with message and files to interaction")
	}
}

type modalRespond struct {
	i   *discordgo.InteractionCreate
	s   *discordgo.Session
	msh *ModalSubmitHandler
}

type Form struct {
	Title             string
	MessageComponents []discordgo.MessageComponent
}

// https://discord.com/developers/docs/interactions/message-components#component-object-component-types
func (r *modalRespond) Form(form Form, cb ModalSubmitCallback) {
	customID := r.msh.AddCallback(cb)

	err := r.s.InteractionRespond(r.i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			CustomID:   customID,
			Content:    form.Title,
			Flags:      discordgo.MessageFlagsEphemeral,
			Components: form.MessageComponents,
		},
	})
	if err != nil {
		log.Error().Err(err).Msg("cannot respond with modal to interaction")
	}
}

type ModalSubmitCallback func(*Responder)

type ModalSubmitHandler struct {
	callbacks map[string]ModalSubmitCallback
}

func NewModalSubmitHandler() *ModalSubmitHandler {
	return &ModalSubmitHandler{
		callbacks: map[string]ModalSubmitCallback{},
	}
}

func (h *ModalSubmitHandler) AddCallback(cb ModalSubmitCallback) string {
	customID := helpers.GetRandomString()
	h.callbacks[customID] = cb
	return customID
}

func (h *ModalSubmitHandler) Start(s *discordgo.Session) {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Interaction.Type != discordgo.InteractionModalSubmit {
			return
		}

		data := i.ModalSubmitData()

		handler, ok := h.callbacks[data.CustomID]
		if !ok {
			return
		}

		handler(NewResponder(i, s, h))

		delete(h.callbacks, data.CustomID)
	})
}
