package config

import "os"

type Config struct {
	SubmitMapsChannelName string
	TestingChannelFormat  string
	SectionName           string
	AppID                 string
	GuildID               string
	Token                 string
}

var CONFIG = Config{
	SubmitMapsChannelName: os.Getenv("SUBMIT_MAPS_CHANNEL_NAME"),
	AppID:                 os.Getenv("APP_ID"),
	GuildID:               os.Getenv("GUILD_ID"),
	SectionName:           os.Getenv("SECTION_NAME"),
	TestingChannelFormat:  os.Getenv("TESTING_CHANNEL_FORMAT"),
	Token:                 os.Getenv("TOKEN"),
}
