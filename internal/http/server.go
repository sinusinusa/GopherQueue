package httpserver

import (
	"context"
	_ "embed"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"GopherQueue/internal/config"
)

type Server struct {
	cfg        *config.Config
	httpServer *http.Server
}

func New(cfg *config.Config) *Server {
	mux := http.NewServeMux()

	s := &Server{cfg: cfg}

	// Go 1.22: метод-ориентированные роуты
	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.HandleFunc("GET /config", s.handleConfig)
	mux.HandleFunc("GET /description", s.handleDescription)

	s.httpServer = &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      logging(mux),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	return s
}

func (s *Server) Start() error {
	log.Printf("http listening on %s", s.httpServer.Addr)
	// http.ErrServerClosed — нормальное завершение
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

//go:embed description
var description []byte

func (s *Server) handleDescription(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write(description)
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	public := s.cfg.PublicCopy()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(public)
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
