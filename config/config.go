package config

type Config struct {
	SubmitMapsChannelID    string
	SubmitMapsChannelName  string
	TestingChannelFormat   string
	TesterSectionChannelID string
	UploadRateLimit        int
	AppID                  string
	GuildID                string
	Token                  string
}

var CONFIG = Config{
	SubmitMapsChannelID:    "1220495375501230131",
	SubmitMapsChannelName:  "submit_maps", // TODO: get channel name by id
	AppID:                  "1220129996471795773",
	GuildID:                "1220129359411810377",
	TesterSectionChannelID: "1224473590989062274",
	TestingChannelFormat:   "mapping_channel_%s",
	Token:                  "",
}
