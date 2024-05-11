package channel

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/config"
)

func createOrGetTextChannel(name string, s *discordgo.Session) (*discordgo.Channel, error) {
	channels, err := s.GuildChannels(config.CONFIG.GuildID)
	if err != nil {
		return nil, fmt.Errorf("cannot list all channels in guild: %w", err)
	}

	for i := range channels {
		if channels[i].Name == name && channels[i].Type == discordgo.ChannelTypeGuildText {
			return channels[i], nil
		}
	}

	ch, err := s.GuildChannelCreate(
		config.CONFIG.GuildID,
		name,
		discordgo.ChannelTypeGuildText,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create %q text channel: %w", name, err)
	}

	return ch, nil
}

func createOrGetCategoryChannel(name string, s *discordgo.Session) (*discordgo.Channel, error) {
	channels, err := s.GuildChannels(config.CONFIG.GuildID)
	if err != nil {
		return nil, fmt.Errorf("cannot list all channels in guild: %w", err)
	}

	for i := range channels {
		if channels[i].Name == name && channels[i].Type == discordgo.ChannelTypeGuildCategory {
			return channels[i], nil
		}
	}

	ch, err := s.GuildChannelCreate(
		config.CONFIG.GuildID,
		name,
		discordgo.ChannelTypeGuildCategory,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create %q category channel: %w", name, err)
	}

	return ch, nil
}
