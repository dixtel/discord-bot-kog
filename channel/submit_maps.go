package channel

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/config"
	"github.com/dixtel/dicord-bot-kog/roles"
)

type submitMapChannel struct {
	raw      *discordgo.Channel
	category *discordgo.Channel
	pos      int
	s        *discordgo.Session
	roles    *roles.BotRoles
}

func createOrGetSubmitMapChannel(s *discordgo.Session, roles *roles.BotRoles) (*submitMapChannel, error) {
	ch, err := createOrGetTextChannel(config.CONFIG.SubmitMapsChannelName, s)
	if err != nil {
		return nil, fmt.Errorf("cannot create or get channel: %w", err)
	}

	category, err := createOrGetCategoryChannel(config.CONFIG.SectionName, s)
	if err != nil {
		return nil, fmt.Errorf("cannot create or get channel: %w", err)
	}

	self := &submitMapChannel{raw: ch, s: s, category: category, pos: 0, roles: roles}

	err = self.UpdateChannel()
	if err != nil {
		return nil, fmt.Errorf("cannot update channel: %w", err)
	}

	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.ChannelID != self.GetID() {
			return
		}

		if m.Author.Bot {
			return
		}

		if len(m.Content) != 0 {
			s.ChannelMessageDelete(m.ChannelID, m.ID)
		}
	})

	return self, nil
}

func (ch *submitMapChannel) GetID() string {
	return ch.raw.ID
}

func (ch *submitMapChannel) UpdateChannel() error {
	updatedChannel, err := ch.s.ChannelEdit(ch.raw.ID, &discordgo.ChannelEdit{
		Name:     ch.raw.Name,
		Position: ch.pos,
		ParentID: ch.category.ID,
		PermissionOverwrites: []*discordgo.PermissionOverwrite{
			{
				ID:   ch.roles.EveryoneRoleID,
				Type: discordgo.PermissionOverwriteTypeRole,
				Deny: (1 << 50) - 1,
				Allow: discordgo.PermissionAddReactions |
					discordgo.PermissionViewChannel |
					discordgo.PermissionSendMessages |
					discordgo.PermissionReadMessageHistory |
					discordgo.PermissionAttachFiles |
					discordgo.PermissionUseSlashCommands,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("cannot update channel %q: %w", ch.raw.Name, err)
	}

	ch.raw = updatedChannel

	return nil
}
