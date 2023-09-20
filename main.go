package main

import (
	"context"
	"os"

	"github.com/holedaemon/lastfm"
	"github.com/holedaemon/microgopster/internal/web"
	"github.com/zikaeroh/ctxlog"
	"go.uber.org/zap"
)

func main() {
	apiKey := os.Getenv("MICROGOPSTER_LAST_API_KEY")
	addr := os.Getenv("MICROGOPSTER_ADDR")
	debug := os.Getenv("MICROGOPSTER_DEBUG") != ""

	logger := ctxlog.New(debug)

	if apiKey == "" {
		logger.Fatal("$MICROGOPSTER_LAST_API_KEY is not set")
	}

	if addr == "" {
		logger.Fatal("$MICROGOPSTER_ADDR is not set")
	}

	lfm, err := lastfm.New(apiKey)
	if err != nil {
		logger.Fatal("error creating last.fm client", zap.Error(err))
	}

	srv, err := web.New(
		web.WithAddr(addr),
		web.WithLastFM(lfm),
	)
	if err != nil {
		logger.Fatal("error creating server", zap.Error(err))
	}

	ctx := ctxlog.WithLogger(context.Background(), logger)
	if err := srv.Run(ctx); err != nil {
		logger.Error("error running server", zap.Error(err))
	}
}
