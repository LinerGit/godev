package main

import (
	"context"
	"fmt"
	"godev/app"
	"os"
	"os/signal"
)

func main() {
	app := app.New()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	err := app.Start(ctx)
	if err != nil {
		fmt.Println("Error starting app:", err)
	}
}
