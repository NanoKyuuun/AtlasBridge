package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/atlasbridge/atlasbridge/internal/app"
	"github.com/atlasbridge/atlasbridge/internal/config"
)

func main() {
	cfg, routes, profiles, err := config.LoadFull()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	application, err := app.New(cfg, routes, profiles)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize app: %v\n", err)
		os.Exit(1)
	}

	sigQuit := make(chan os.Signal, 1)
	signal.Notify(sigQuit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		_ = application.Run()
	}()

	select {
	case <-sigQuit:
		application.Shutdown()
	case <-func() chan struct{} {
		ch := make(chan struct{})
		go func() {
			application.WaitQuit()
			close(ch)
		}()
		return ch
	}():
	case err := <-application.ErrCh():
		fmt.Fprintf(os.Stderr, "application error: %v\n", err)
		application.Shutdown()
		os.Exit(1)
	}
}
