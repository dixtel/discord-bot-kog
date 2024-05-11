package command

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/helpers"
	"github.com/dixtel/dicord-bot-kog/models"
	"github.com/dixtel/dicord-bot-kog/roles"
)

func getAttachmentFromOption(i *discordgo.InteractionCreate, optionName string) *discordgo.MessageAttachment {
	option := helpers.First(
		i.ApplicationCommandData().Options,
		func(val *discordgo.ApplicationCommandInteractionDataOption) bool {
			return val.Name == optionName
		},
	)
	if option == nil {
		return nil
	}

	// TODO: prevent from panic
	return i.ApplicationCommandData().Resolved.Attachments[(*option).Value.(string)]
}

func getUserFromOption(s *discordgo.Session, i *discordgo.InteractionCreate, optionName string) *discordgo.User {
	option := helpers.First(
		i.ApplicationCommandData().Options,
		func(val *discordgo.ApplicationCommandInteractionDataOption) bool {
			return val.Name == optionName
		},
	)
	if option == nil {
		return nil
	}

	return (*option).UserValue(s)
}

func getUsername(i *discordgo.InteractionCreate) string {
	if i.Member != nil && i.Member.Nick != ""  {
		return i.Member.Nick
	} 

	if i.Member.User != nil && i.Member.User.Username != ""  {
		return i.Member.User.Username
	} 

	return "<undefined>"
}

func mentionUser(i *discordgo.InteractionCreate) string {
	return fmt.Sprintf("<@%s>", i.Member.User.ID)
}

type ApproveORDecline = string
var (
	ApproveORDecline_Approve = "approve"
	ApproveORDecline_Decline = "decline"
)

func approveOrDecline(
	database *models.Database,
	botRoles *roles.BotRoles,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	action ApproveORDecline,
) error {
	interaction, err := FromDiscordInteraction(s, i)
	if err != nil {
		return fmt.Errorf("cannot create interaction: %w", err)
	}

	_, err = database.CreateOrGetUser(i.Member.User.Username, i.Member.User.ID)
	if err != nil {
		return fmt.Errorf("cannot create or get an user: %w", err)
	}

	isTestingChannel, err := database.IsTestingChannel(i.ChannelID)
	if err != nil {
		return fmt.Errorf("cannot check if current channel is a testing channel: %w", err)
	}

	if !isTestingChannel {
		return interaction.SendMessage(
			"This is not a channel for testing",
			InteractionMessageType_Private,
		)
	}

	if !botRoles.HasMapAcceptorRole(i.Member) && !botRoles.HasMapTesterRole(i.Member) {
		return interaction.SendMessage(
			"You don't have role 'Tester' or 'Map Acceptor' to approve the map.",
			InteractionMessageType_Private,
		)
	}

	m, err := database.GetLastUploadedMapByChannelID(i.ChannelID)
	if err != nil {
		return fmt.Errorf("cannot get last uploaded map by channel id: %w", err)
	}

	if m.Status != models.MapStatus_Testing {
		return interaction.SendMessage(
			fmt.Sprintf("Cannot %s the map. Current map status is %q.", action, m.Status),
			InteractionMessageType_Private,
		)
	}

	data, err := database.GetTestingChannelData(m.ID)
	if err != nil {
		return fmt.Errorf("cannot get testing channel data %w", err)
	}

	if _, ok := data.ApprovedBy[i.Member.User.ID]; ok {
		return interaction.SendMessage(
			"You already approved this map",
			InteractionMessageType_Private,
		)
	}

	if _, ok := data.DeclinedBy[i.Member.User.ID]; ok {
		return interaction.SendMessage(
			"You already declined this map",
			InteractionMessageType_Private,
		)
	}

	if action == ApproveORDecline_Approve {
		data.ApprovedBy[i.Member.User.ID] = struct{}{}
	} else {
		data.DeclinedBy[i.Member.User.ID] = struct{}{}
	}

	err = database.UpdateTestingChannelData(m.ID, data)
	if err != nil {
		return fmt.Errorf("cannot update testing channel data %w", err)
	}

	return interaction.SendMessage(
		fmt.Sprintf(
			"Map was %sd by tester %s. (%v approvals / %v declines)",
			action, mentionUser(i), len(data.ApprovedBy), len(data.DeclinedBy),
		),
		InteractionMessageType_Public,
	)
}