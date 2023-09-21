package web

type Option func(*Server)

func WithAddr(addr string) Option {
	return func(s *Server) {
		s.addr = addr
	}
}

func WithAPIKey(key string) Option {
	return func(s *Server) {
		s.apiKey = key
	}
}
