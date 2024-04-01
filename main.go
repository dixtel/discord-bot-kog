package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/bot"
	"github.com/dixtel/dicord-bot-kog/config"
	"github.com/dixtel/dicord-bot-kog/models"
	"github.com/dixtel/dicord-bot-kog/roles"
	"github.com/glebarez/sqlite"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func main() {
	gormDb, err := gorm.Open(sqlite.Open("sqlite.db"), &gorm.Config{})
	if err != nil {
		log.Error().Err(err).Msg("cannot connect to database")
		return
	}

	err = gormDb.AutoMigrate(
		&models.User{},
		&models.Map{},
		&models.Role{},
		&models.TestingChannel{},
	)
	if err != nil {
		log.Error().Err(err).Msg("cannot migrate models")
		return
	}

	db := &models.Database{
		DB: gormDb.Debug(),
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + config.CONFIG.Token)
	if err != nil {
		log.Error().Err(err).Msg("error creating Discord session")
		return
	}

	botRoles, err := roles.SetupRoles(dg)
	if err != nil {
		log.Error().Err(err).Msg("cannot setup roles")
		return
	}

	defer bot.SetupCommands(dg, db, botRoles)()

	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Info().Msg("Bot is up!")
	})

	// Just like the ping pong example, we only care about receiving message
	// events in this example.
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentMessageContent

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		log.Error().Err(err).Msg("error opening connection")
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	log.Info().Msg("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}
