package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type App struct {
	router http.Handler
	rdb    *redis.Client
}

func New() *App {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	app := &App{
		router: loadRoutes(),
		rdb:    rdb,
	}
	return app
}
func (a *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    ":3000",
		Handler: a.router,
	}

	err := a.rdb.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	defer func() {
		if err := a.rdb.Close(); err != nil {
			fmt.Printf("Error closing Redis client: %v\n", err)
		}
	}()

	fmt.Println("Connected to Redis successfully")

	ch := make(chan error, 1)
	go func() {
		err = server.ListenAndServe()
		if err != nil {
			ch <- fmt.Errorf("failed start server")
		}

	}()
	select {
	case err = <-ch:
		return err
	case <-ctx.Done():
		fmt.Println("Shutting down server...")
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		return server.Shutdown(timeout)
	}

}
