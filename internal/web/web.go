package web

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/holedaemon/lastfm"
	"github.com/zikaeroh/ctxlog"
	"go.uber.org/zap"
)

//go:embed static
var static embed.FS

var staticDir fs.FS

func init() {
	var err error
	staticDir, err = fs.Sub(static, "static")
	if err != nil {
		panic(err)
	}
}

const (
	version   = "0"
	userAgent = "microgopster/v" + version + " (https://github.com/holedaemon/microgopster)"
)

type Server struct {
	addr   string
	apiKey string

	cli    *http.Client
	lastfm *lastfm.Client
}

func New(opts ...Option) (*Server, error) {
	s := &Server{
		cli: &http.Client{
			Timeout: time.Second * 10,
		},
	}

	for _, o := range opts {
		o(s)
	}

	if s.apiKey == "" {
		return nil, fmt.Errorf("web: missing last.fm api key")
	}

	if s.addr == "" {
		return nil, fmt.Errorf("web: missing addr")
	}

	lfm, err := lastfm.New(s.apiKey, lastfm.UserAgent(userAgent))
	if err != nil {
		return nil, err
	}

	s.lastfm = lfm

	return s, nil
}

func (s *Server) Run(ctx context.Context) error {
	r := chi.NewMux()

	r.Use(recoverer)

	logger := ctxlog.FromContext(ctx)
	r.Use(requestLogger(logger))

	r.Post("/", s.index)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		respondError(w, r, http.StatusNotFound, http.StatusText(http.StatusNotFound))
	})

	srv := &http.Server{
		Addr:        s.addr,
		Handler:     r,
		BaseContext: func(l net.Listener) context.Context { return ctx },
	}

	go func() {
		<-ctx.Done()
		if err := srv.Shutdown(context.Background()); err != nil {
			ctxlog.Error(ctx, "error shutting down server", zap.Error(err))
			return
		}
	}()

	ctxlog.Info(ctx, "web server listening", zap.String("addr", srv.Addr))
	return srv.ListenAndServe()
}
