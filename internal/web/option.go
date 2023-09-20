package web

import (
	"net/http"

	"github.com/holedaemon/lastfm"
)

type Option func(*Server)

func WithAddr(addr string) Option {
	return func(s *Server) {
		s.Addr = addr
	}
}

func WithLastFM(lfm *lastfm.Client) Option {
	return func(s *Server) {
		s.LastFM = lfm
	}
}

func WithClient(cli *http.Client) Option {
	return func(s *Server) {
		s.HTTP = cli
	}
}
