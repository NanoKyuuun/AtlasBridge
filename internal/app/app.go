package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/smart-ai-proxy/smart-ai-proxy/internal/config"
	"github.com/smart-ai-proxy/smart-ai-proxy/internal/forwarder"
	"github.com/smart-ai-proxy/smart-ai-proxy/internal/observability"
	"github.com/smart-ai-proxy/smart-ai-proxy/internal/runtime"
	"github.com/smart-ai-proxy/smart-ai-proxy/internal/server"
	"github.com/smart-ai-proxy/smart-ai-proxy/internal/startup"
	"github.com/smart-ai-proxy/smart-ai-proxy/internal/tray"
)

type App struct {
	cfg       *config.Config
	server    *http.Server
	listener  net.Listener
	routes    *config.RoutesConfig
	profiles  *config.ProfilesConfig
	state     *runtime.State
	tray      *tray.Tray
	quitCh    chan struct{}
}

func New(cfg *config.Config, routes *config.RoutesConfig, profiles *config.ProfilesConfig) (*App, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("port conflict on %s: %w (another process may be using port %d)", addr, err, cfg.Server.Port)
	}

	mode := runtime.Mode(cfg.App.Mode)
	state := runtime.NewState(mode)

	fwd, err := forwarder.New(cfg.Downstream.BaseURL, cfg.Downstream.TimeoutSeconds)
	if err != nil {
		log.Printf("WARNING: failed to create forwarder: %v", err)
	}

	obsLogger := observability.NewLogger(observability.DefaultMaxEntries)

	deps := &server.ServerDeps{
		Config:       cfg,
		Routes:       routes,
		Profiles:     profiles,
		Fwd:          fwd,
		ObsLogger:    obsLogger,
		RuntimeState: state,
	}

	srv := server.New(deps)

	return &App{
		cfg:      cfg,
		server:   srv,
		listener: ln,
		routes:   routes,
		profiles: profiles,
		state:    state,
		quitCh:   make(chan struct{}),
	}, nil
}

func (a *App) Run() error {
	if err := startup.Init(); err != nil {
		return fmt.Errorf("another instance is already running: %w", err)
	}

	a.tray = tray.New(a.cfg, a.state)

	go a.tray.Run(func() {
		a.state.SetStatus(runtime.StatusStopped)
		a.Shutdown()
		close(a.quitCh)
	})

	shouldAutoStart := a.cfg.App.Mode == string(runtime.ModeAlwaysOn) && a.cfg.Startup.StartProxyOnAppLaunch
	if shouldAutoStart {
		if err := a.state.Start(); err != nil {
			log.Printf("auto-start failed: %v", err)
		}
	}

	addr := fmt.Sprintf("%s:%d", a.cfg.Server.Host, a.cfg.Server.Port)
	log.Printf("Smart AI Proxy v%s starting", server.Version)
	log.Printf("Listening on http://%s", addr)
	log.Printf("API endpoint: http://%s%s", addr, a.cfg.Server.APIBasePath)
	log.Printf("Admin UI:     http://%s%s", addr, a.cfg.Server.AdminPath)
	log.Printf("Health check: http://%s/health", addr)
	log.Printf("Downstream:   %s", a.cfg.Downstream.BaseURL)
	log.Printf("Privacy mode: %s", a.cfg.Logging.PrivacyMode)
	log.Printf("Auto routing: %v", a.cfg.Routing.AutoRouting)

	if err := a.server.Serve(a.listener); err != nil && err != http.ErrServerClosed {
		a.state.SetError(err.Error())
		a.tray.Update()
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

func (a *App) WaitQuit() {
	<-a.quitCh
}

func (a *App) Shutdown() {
	log.Println("Smart AI Proxy shutting down...")
	a.state.SetStatus(runtime.StatusStopped)
	if a.tray != nil {
		a.tray.Update()
	}
	if err := a.server.Shutdown(context.Background()); err != nil {
		log.Printf("shutdown error: %v", err)
	}
	startup.ReleaseLock()
	log.Println("Smart AI Proxy stopped.")
}
