package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mvisonneau/slack-git-compare/pkg/slack"

	"github.com/felixge/httpsnoop"
	"github.com/gorilla/mux"
	"github.com/heptiolabs/healthcheck"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func getHTTPServer(listenAddress string, s slack.Slack) *http.Server {
	router := mux.NewRouter()
	loggerRouter := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := httpsnoop.CaptureMetrics(router, w, r)
		log.WithFields(
			log.Fields{
				"client":        r.RemoteAddr,
				"path":          r.RequestURI,
				"user_agent":    r.UserAgent(),
				"method":        r.Method,
				"proto":         r.Proto,
				"host":          r.Host,
				"code":          m.Code,
				"duration":      m.Duration,
				"bytes_written": m.Written,
			},
		).Debug()
	})

	// health endpoints
	health := healthcheck.NewHandler()
	router.HandleFunc("/health/live", health.LiveEndpoint)
	router.HandleFunc("/health/ready", health.ReadyEndpoint)

	// main endpoint
	router.HandleFunc("/slack/slash", s.SlashHandler)
	router.HandleFunc("/slack/modal", s.ModalHandler)
	router.HandleFunc("/slack/select", s.SelectHandler)

	return &http.Server{
		Addr:    listenAddress,
		Handler: loggerRouter,
	}
}

// Run launches the exporter
func Run(cliContext *cli.Context) (int, error) {
	cfg := configure(cliContext)
	ctx := context.TODO()
	s, err := slack.New(ctx, cfg.Slack)
	if err != nil {
		log.Fatal(err)
	}

	// Graceful shutdowns
	onShutdown := make(chan os.Signal, 1)
	signal.Notify(onShutdown, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)

	// HTTP server
	srv := getHTTPServer(cfg.ListenAddress, s)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	log.WithFields(
		log.Fields{
			"listen-address": cfg.ListenAddress,
		},
	).Info("http server started")

	<-onShutdown
	log.Info("received signal, attempting to gracefully exit..")

	httpServerContext, forceHTTPServerShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer forceHTTPServerShutdown()

	if err := srv.Shutdown(httpServerContext); err != nil {
		log.WithError(err).Fatalf("metrics server shutdown failed")
	}

	log.Info("stopped!")
	return 0, nil
}
