package webserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/dixtel/dicord-bot-kog/config"
	"github.com/rs/zerolog/log"
)

func errorResponse(w http.ResponseWriter, status int, msg string, err error) {
	w.WriteHeader(status)
	w.Write([]byte(msg))
    if err != nil {
        log.Error().Err(err).Msg("web server error")
    }
}

type FeedbackPost struct {
	Username string `json:"username"`
	Map      string `json:"map"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
	Text     string `json:"text"`
}

func mapFeedbackHandler(s *discordgo.Session) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errorResponse(w, http.StatusBadRequest, "expects POST request", nil)
			return
		}

		data := &FeedbackPost{}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			errorResponse(w, http.StatusBadRequest, "cannot read body", err)
			return
		}

		err = json.Unmarshal(body, data)
		if err != nil {
			errorResponse(w, http.StatusBadRequest, "cannot decode body", err)
			return
		}

		_, err = s.ChannelMessageSend(
			config.CONFIG.MapFeedbackChannelID,
			fmt.Sprintf(
                "New map feedback from %s\nMap: %s\nCords: x=%v, y=%v\nMessage: %s",
                data.Username,
                data.Map,
                data.X,
                data.Y,
                data.Text,
            ),
        )
        if err != nil {
			errorResponse(w, http.StatusInternalServerError, "cannot handle request", err)
			return
		}
	}

}

func Run(s *discordgo.Session) *http.Server {
	log.Info().Msg("web server is running on http://localhost:8080")
	srv := &http.Server{Addr: ":8080"}

	http.HandleFunc("/api/v1/event/map-feedback", mapFeedbackHandler(s))

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Err(err).Msg("web server error")
		}
	}()

	return srv
}
