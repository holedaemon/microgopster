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

type Server struct {
	Addr   string
	HTTP   *http.Client
	LastFM *lastfm.Client
}

func New(opts ...Option) (*Server, error) {
	s := &Server{}

	for _, o := range opts {
		o(s)
	}

	if s.HTTP == nil {
		s.HTTP = &http.Client{
			Timeout: time.Second * 10,
		}
	}

	if s.LastFM == nil {
		return nil, fmt.Errorf("web: missing lastfm client")
	}

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
		Addr:        s.Addr,
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
