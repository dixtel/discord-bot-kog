package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/bot"
	"github.com/dixtel/dicord-bot-kog/channel"
	"github.com/dixtel/dicord-bot-kog/config"
	"github.com/dixtel/dicord-bot-kog/models"
	"github.com/dixtel/dicord-bot-kog/roles"
	"github.com/dixtel/dicord-bot-kog/webserver"
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
		&models.TestingChannel{},
		&models.BannedUserFromSubmission{},
	)
	if err != nil {
		log.Error().Err(err).Msg("cannot migrate models")
		return
	}

	db := models.NewDatabase(gormDb.Debug())

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

	channelManager, err := channel.NewChannelManager(dg, botRoles)
	if err != nil {
		log.Error().Err(err).Msg("cannot create channel manager")
		return
	}


	// defer bot.SetupBot(dg, db, botRoles, channelManager)()

	cleanup, err := bot.SetupBot(dg, db, botRoles, channelManager)
	if err != nil {
		log.Error().Err(err).Msg("cannot setup bot v2")
		return
	}

	defer cleanup()

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

	srv := webserver.Run(dg)

	// Wait here until CTRL-C or other term signal is received.
	log.Info().Msg("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc



	// Cleanly close down the Discord session.
	err = dg.Close()
	if err != nil {
		log.Err(err).Msg("cannot close discord bot")
	}

	if err := srv.Shutdown(context.TODO()); err != nil {
        panic(err) // failure/timeout shutting down the server gracefully
    }
}
