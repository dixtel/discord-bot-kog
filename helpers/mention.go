package helpers

import "fmt"

func MentionUser(userID string) string {
	return fmt.Sprintf("<@%s>", userID)
}

func MentionChannel(channelID string) string {
	return fmt.Sprintf("<#%s>", channelID)
}

func MentionRole(roleID string) string {
	return fmt.Sprintf("<@&%s>", roleID)
}