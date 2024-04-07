package roles

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/config"
	"github.com/dixtel/dicord-bot-kog/helpers"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type BotRoles struct {
	MapTester   *discordgo.Role
	MapAcceptor *discordgo.Role
}

func (br *BotRoles) HasMapTesterRole(member *discordgo.Member) bool {
	log.Debug().
		Array("member roles", zerolog.Arr().Interface(member.Roles)).
		Msgf("has member have access to role 'Map Tester' id: %s", br.MapTester.ID)

	r := helpers.GetFromArr(member.Roles, func(v string) bool {
		return v == br.MapTester.ID
	})
	return r != nil
}

func (br *BotRoles) HasMapAcceptorRole(member *discordgo.Member) bool {
	log.Debug().
		Array("member roles", zerolog.Arr().Interface(member.Roles)).
		Msgf("has member have access to role 'Map Acceptor' id: %s", br.MapTester.ID)

	r := helpers.GetFromArr(member.Roles, func(v string) bool {
		return v == br.MapAcceptor.ID
	})
	return r != nil
}

func createOrGetRole(s *discordgo.Session, name string) (*discordgo.Role, error) {
	currentRoles, err := s.GuildRoles(config.CONFIG.GuildID)
	if err != nil {
		return nil, fmt.Errorf("cannot get guild roles")
	}

	role := helpers.GetFromArr(currentRoles, func(r *discordgo.Role) bool {
		return r.Name == name
	})

	if role != nil {
		return *role, nil
	}

	newRole, err := s.GuildRoleCreate(config.CONFIG.GuildID, &discordgo.RoleParams{
		Name:        name,
		Color:       helpers.ToPtr(123),
		Hoist:       nil,
		Permissions: nil,
		Mentionable: helpers.ToPtr(true),
	})
	if err != nil {
		return nil, fmt.Errorf("cannot create %q role: %w", name, err)
	}

	return newRole, nil
}

func SetupRoles(s *discordgo.Session) (*BotRoles, error) {
	roles := &BotRoles{}

	r, err := createOrGetRole(s, "Map Tester")
	if err != nil {
		return nil, err
	}

	roles.MapTester = r

	r, err = createOrGetRole(s, "Map Acceptor")
	if err != nil {
		return nil, err
	}

	roles.MapAcceptor = r

	return roles, nil
}
