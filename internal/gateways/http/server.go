package http

import (
	"context"
	"fmt"
	"homework/internal/usecase"
	"log"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/gin-gonic/gin"
)

type Server struct {
	host   string
	port   uint16
	router *gin.Engine
}

type UseCases struct {
	Event  *usecase.Event
	Sensor *usecase.Sensor
	User   *usecase.User
}

func NewServer(useCases UseCases, options ...func(*Server)) *Server {
	r := gin.Default()
	setupRouter(r, useCases, NewWebSocketHandler(useCases))

	s := &Server{router: r, host: "localhost", port: 8080}
	for _, o := range options {
		o(s)
	}

	return s
}

func WithHost(host string) func(*Server) {
	return func(s *Server) {
		s.host = host
	}
}

func WithPort(port uint16) func(*Server) {
	return func(s *Server) {
		s.port = port
	}
}

func (s *Server) Run(ctx context.Context) error {
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.host, s.port),
		Handler: s.router,
	}
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return server.ListenAndServe()
	})
	_ = eg.Wait()

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, time.Second*2)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}
	return s.router.Run(fmt.Sprintf("%s:%d", s.host, s.port))
}
