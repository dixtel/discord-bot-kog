package helpers

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/config"
)

func GetEveryoneRole(s *discordgo.Session) (string, error) {
	// https://discord.com/developers/docs/topics/permissions#permissions-bitwise-permission-flags
	// https://discord.com/developers/docs/topics/permissions#permissions
	// https://stackoverflow.com/a/60093794/10300644
	guildRoles, err := s.GuildRoles(config.CONFIG.GuildID)
	if err != nil {
		return "", fmt.Errorf("cannot get guild roles: %w", err)
	}

	everyoneRole := GetFromArr(guildRoles, func(r *discordgo.Role) bool {
		return r.Name == "@everyone"
	})

	if everyoneRole == nil {
		return "", fmt.Errorf("cannot get 'everyone' role")
	}

	return (*everyoneRole).ID, nil
}