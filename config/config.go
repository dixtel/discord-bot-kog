package config

type Config struct {
	SubmitMapsChannelID  string
	MappingChannelFormat string
	UploadRateLimit      int
	AppID                string
	GuildID              string
	Token string 
}

var CONFIG = Config{
	SubmitMapsChannelID:  "1220495375501230131",
	AppID:                "1220129996471795773",
	GuildID:              "1220129359411810377",
	MappingChannelFormat: "mapping_channel_%s",
	Token: "MTIyMDEyOTk5NjQ3MTc5NTc3Mw.GPGT7q.ZgFpAMzEAmTs_xqf6nD7T8vNWVt5yESUTG6LfI",
}