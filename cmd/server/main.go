package main

import (
	"context"
	"errors"
	"fmt"
	"homework/internal/usecase"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/jackc/pgx/v5/pgxpool"

	httpGateway "homework/internal/gateways/http"
	eventRepository "homework/internal/repository/event/postgres"
	sensorRepository "homework/internal/repository/sensor/postgres"
	userRepository "homework/internal/repository/user/postgres"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	config, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("can't parse pgxpool config")
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("can't create new pool")
	}
	defer pool.Close()

	er := eventRepository.NewEventRepository(pool)
	sr := sensorRepository.NewSensorRepository(pool)
	ur := userRepository.NewUserRepository(pool)
	sor := userRepository.NewSensorOwnerRepository(pool)

	useCases := httpGateway.UseCases{
		Event:  usecase.NewEvent(er, sr),
		Sensor: usecase.NewSensor(sr),
		User:   usecase.NewUser(ur, sor, sr),
	}

	var host string
	var port int
	h, ok := os.LookupEnv("HTTP_HOST")
	if !ok {
		h = "localhost"
	}
	host = h
	p, ok := os.LookupEnv("HTTP_PORT")
	if !ok {
		p = "8080"
	}
	port, err = strconv.Atoi(p)
	if err != nil {
		port = 8080
	}

	sigQuit := make(chan os.Signal, 1)
	signal.Notify(sigQuit, syscall.SIGTERM, syscall.SIGINT)
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		s := <-sigQuit
		_, err := fmt.Printf("capturet signal: %v", s)
		return err
	})

	r := httpGateway.NewServer(useCases, httpGateway.WithHost(host), httpGateway.WithPort(uint16(port)))

	eg.Go(func() error {
		return r.Run(ctx)
	})

	if err := eg.Wait(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("error during server shutdown: %v", err)
	}
}
