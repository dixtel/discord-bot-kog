package webserver

import (
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

func handle(w http.ResponseWriter, r *http.Request) {
    println("GET", r.URL.RawPath)

    io.WriteString(w, "hello world\n")
}

func Run() *http.Server {
    println("web server is running on http://localhost:8080")
    srv := &http.Server{Addr: ":8080"}

    http.HandleFunc("/", handle)

    go func() {
        if err := srv.ListenAndServe(); err != http.ErrServerClosed {
            log.Err(err).Msg("web server error")
        }
    }()

    return srv
}
