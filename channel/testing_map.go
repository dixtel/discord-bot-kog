package channel

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/config"
	"github.com/dixtel/dicord-bot-kog/roles"
	"github.com/dixtel/dicord-bot-kog/twmap"
)

type testingMapChannel struct {
	raw          *discordgo.Channel
	category     *discordgo.Channel
	pos          int
	s            *discordgo.Session
	roles        *roles.BotRoles
	mapCreatorID string
}

func createTestingMapChannel(
	s *discordgo.Session,
	mapFilename string,
	roles *roles.BotRoles,
	mapCreatorID string,
) (
	*testingMapChannel,
	error,
) {
	channelName := fmt.Sprintf(
		config.CONFIG.TestingChannelFormat,
		twmap.RemoveMapFileExtension(mapFilename),
	)
	
	ch, err := createTextChannel(channelName, s)
	if err != nil {
		return nil, fmt.Errorf("cannot create: %w", err)
	}

	category, err := createOrGetCategoryChannel(config.CONFIG.SectionName, s)
	if err != nil {
		return nil, fmt.Errorf("cannot create or get channel: %w", err)
	}

	self := &testingMapChannel{raw: ch, s: s, category: category, pos: 1, roles: roles, mapCreatorID: mapCreatorID}

	err = self.UpdateChannel()
	if err != nil {
		return nil, fmt.Errorf("cannot update channel: %w", err)
	}

	return self, nil
}

func (ch *testingMapChannel) GetName() string {
	return ch.raw.Name
}

func (ch *testingMapChannel) GetID() string {
	return ch.raw.ID
}

func (ch *testingMapChannel) UpdateChannel() error {
	updatedChannel, err := ch.s.ChannelEdit(ch.raw.ID, &discordgo.ChannelEdit{
		Name:     ch.raw.Name,
		Position: ch.pos,
		ParentID: ch.category.ID,
		PermissionOverwrites: []*discordgo.PermissionOverwrite{
			{
				ID:   ch.roles.EveryoneRoleID,
				Type: discordgo.PermissionOverwriteTypeRole,
				Deny: (1 << 50) - 1,
			},
			{
				ID:   ch.roles.MapAcceptor.ID,
				Type: discordgo.PermissionOverwriteTypeRole,
				Deny: (1 << 50) - 1,
				Allow: discordgo.PermissionAddReactions |
					discordgo.PermissionViewChannel |
					discordgo.PermissionReadMessageHistory |
					discordgo.PermissionSendMessages |
					discordgo.PermissionAttachFiles |
					discordgo.PermissionUseSlashCommands,
			},
			{
				ID:   ch.roles.MapTester.ID,
				Type: discordgo.PermissionOverwriteTypeRole,
				Deny: (1 << 50) - 1,
				Allow: discordgo.PermissionAddReactions |
					discordgo.PermissionViewChannel |
					discordgo.PermissionReadMessageHistory |
					discordgo.PermissionSendMessages |
					discordgo.PermissionAttachFiles |
					discordgo.PermissionUseSlashCommands,
			},
			{
				ID:   ch.mapCreatorID,
				Type: discordgo.PermissionOverwriteTypeRole,
				Deny: (1 << 50) - 1,
				Allow: discordgo.PermissionAddReactions |
					discordgo.PermissionViewChannel |
					discordgo.PermissionReadMessageHistory |
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
