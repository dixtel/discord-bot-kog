package channel

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/config"
	"github.com/dixtel/dicord-bot-kog/helpers"
)

type SubmitMapChannel struct {
	raw      *discordgo.Channel
	category *discordgo.Channel
	pos      int
	s        *discordgo.Session
}

func CreateSubmitMapChannel(s *discordgo.Session) (*SubmitMapChannel, error) {
	ch, err := createOrGetTextChannel(config.CONFIG.SubmitMapsChannelName, s)
	if err != nil {
		return nil, fmt.Errorf("cannot crate or get channel: %w", err)
	}

	category, err := createOrGetCategoryChannel(config.CONFIG.SectionName, s)
	if err != nil {
		return nil, fmt.Errorf("cannot crate or get channel: %w", err)
	}

	return &SubmitMapChannel{raw: ch, s: s, category: category, pos: 0}, nil
}

func (ch *SubmitMapChannel) UpdateChannel() error {
	everyoneRoleID, err := helpers.GetEveryoneRole(ch.s)
	if err != nil {
		return fmt.Errorf("cannot get '@everyone' role")
	}

	updatedChannel, err := ch.s.ChannelEdit(ch.raw.ID, &discordgo.ChannelEdit{
		Name:     ch.raw.Name,
		Position: ch.pos,
		ParentID: ch.category.ID,
		PermissionOverwrites: []*discordgo.PermissionOverwrite{
			{
				ID:   everyoneRoleID,
				Type: discordgo.PermissionOverwriteTypeRole,
				Deny: (1 << 50) - 1,
				Allow: discordgo.PermissionAddReactions |
					discordgo.PermissionViewChannel |
					discordgo.PermissionViewChannel |
					discordgo.PermissionSendMessages |
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
