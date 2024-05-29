package channel

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/roles"
)

type ChannelManager struct {
	s                *discordgo.Session
	roles            *roles.BotRoles
	submitMapChannel *submitMapChannel
}

func NewChannelManager(s *discordgo.Session, roles *roles.BotRoles) (*ChannelManager, error) {
	submitMapChannel, err := createOrGetSubmitMapChannel(s, roles)
	if err != nil {
		return nil, fmt.Errorf("cannot create or get submission channel: %w", err)
	}

	return &ChannelManager{s, roles, submitMapChannel}, nil
}

func (m *ChannelManager) CreateTestingMapChannel(
	mapFilename string,
	mapCreatorID string,
) (*testingMapChannel, error) {
	return createTestingMapChannel(m.s, mapFilename, m.roles, mapCreatorID)
}

func (m *ChannelManager) GetSubmitMapChannelID() (string) {
	return m.submitMapChannel.GetID()
}
