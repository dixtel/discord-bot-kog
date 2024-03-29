package config

type Config struct {
	SubmitMapsChannelID   string
	SubmitMapsChannelName string
	TestingChannelFormat  string
	UploadRateLimit       int
	AppID                 string
	GuildID               string
	Token                 string
}

var CONFIG = Config{
	SubmitMapsChannelID:   "1220495375501230131",
	SubmitMapsChannelName: "submit_maps", // TODO: get channel name by id
	AppID:                 "1220129996471795773",
	GuildID:               "1220129359411810377",
	TestingChannelFormat:  "mapping_channel_%s",
	Token:                 "MTIyMDEyOTk5NjQ3MTc5NTc3Mw.G0fYEp.qko1VzHPyL5GyMl7or-iKClheHQUcQhUjkV3hQ",
}
