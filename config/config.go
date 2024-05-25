package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	SubmitMapsChannelName     string
	TestingChannelFormat      string
	SectionName               string
	AppID                     string
	GuildID                   string
	Token                     string
	MinimumMapApprovalsNumber int
}

var CONFIG = Config{
	SubmitMapsChannelName:     orDefaultString(os.Getenv("SUBMIT_MAPS_CHANNEL_NAME"), "submit_maps"),
	AppID:                     getEnvOrPanic("APP_ID"),
	GuildID:                   getEnvOrPanic("GUILD_ID"),
	SectionName:               orDefaultString(os.Getenv("SUBMIT_MAPS_CHANNEL_NAME"), "tester section"),
	TestingChannelFormat:      orDefaultString(os.Getenv("TESTING_CHANNEL_FORMAT"), "mapping_channel_%s"),
	Token:                     getEnvOrPanic("TOKEN"),
	MinimumMapApprovalsNumber: orDefaultInt(os.Getenv("MINIMUM_MAP_APPROVALS_NUMBER"), 3),
}

func orDefaultInt(value string, def int) int {
	if value == "" {
		return def
	}

	v, err := strconv.Atoi(value)
	if err != nil {
		panic(err.Error())
	}

	return v
}

func orDefaultString(value string, def string) string {
	if value == "" {
		return def
	}

	return value
}

func getEnvOrPanic(name string) string {
	v := os.Getenv(name)
	if v == "" {
		panic(fmt.Sprint(name, "must be non empty"))
	}

	return v
}
