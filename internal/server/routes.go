package server

import "net/http"

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthz)
	mux.HandleFunc("GET /readyz", s.readyz)
	for _, m := range s.modules {
		m.Register(mux)
	}
	return recoverer(s.log)(requestLogger(s.log)(requestID(mux)))
}
