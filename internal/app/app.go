package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/atlasbridge/atlasbridge/internal/config"
	"github.com/atlasbridge/atlasbridge/internal/forwarder"
	"github.com/atlasbridge/atlasbridge/internal/observability"
	"github.com/atlasbridge/atlasbridge/internal/runtime"
	"github.com/atlasbridge/atlasbridge/internal/security"
	"github.com/atlasbridge/atlasbridge/internal/server"
	"github.com/atlasbridge/atlasbridge/internal/startup"
	"github.com/atlasbridge/atlasbridge/internal/tray"
)

type App struct {
	cfg           *config.Config
	server        *http.Server
	listener      net.Listener
	routes        *config.RoutesConfig
	profiles      *config.ProfilesConfig
	state         *runtime.State
	tray          *tray.Tray
	quitCh        chan struct{}
	errCh         chan error
	store         *server.StateStore
	configService *server.ConfigService
}

func effectiveHost(cfg *config.Config) string {
	config.EnforceNetworkInvariants(cfg)
	return cfg.Server.Host
}

func New(cfg *config.Config, routes *config.RoutesConfig, profiles *config.ProfilesConfig) (*App, error) {
	host := effectiveHost(cfg)
	if host != cfg.Server.Host {
		log.Printf("SECURITY: forcing bind to 127.0.0.1 (allow_lan_access=false, bind_localhost_only=true)")
		cfg.Server.Host = host
	}
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

	retentionDays := cfg.Logging.RetentionDays
	if retentionDays <= 0 {
		retentionDays = 7
	}

	snap := &server.Snapshot{
		Config:    *cfg,
		Routes:    *routes,
		Profiles:  *profiles,
		Forwarder: fwd,
		Version:   1,
		CreatedAt: time.Now(),
	}
	store := server.NewStateStore(snap)
	cs := server.NewConfigService(store)

	deps := &server.ServerDeps{
		Store:         store,
		ConfigService: cs,
		ObsLogger:     obsLogger,
		RuntimeState:  state,
	}

	srv := server.New(deps)

	return &App{
		cfg:           cfg,
		server:        srv,
		listener:      ln,
		routes:        routes,
		profiles:      profiles,
		state:         state,
		quitCh:        make(chan struct{}),
		errCh:         make(chan error, 1),
		store:         store,
		configService: cs,
	}, nil
}

func (a *App) Run() error {
	if err := startup.Init(); err != nil {
		return fmt.Errorf("another instance is already running: %w", err)
	}

	if a.cfg.Security.AdminAuthEnabled {
		rawToken, err := security.EnsureToken(&a.cfg.Security.AdminTokenHash)
		if err != nil {
			log.Printf("WARNING: failed to generate admin token: %v", err)
		} else if rawToken != "" {
			if err := config.Save(a.cfg); err != nil {
				log.Printf("WARNING: failed to save config with admin token: %v", err)
			}
			tokenPath := config.TokenFilePath()
			if err := os.WriteFile(tokenPath, []byte(rawToken), 0o600); err != nil {
				log.Printf("WARNING: failed to write token file for CLI access: %v", err)
			}
			fmt.Fprint(os.Stdout, "\n")
			fmt.Fprintf(os.Stdout, "  ADMIN TOKEN: %s\n", rawToken)
			fmt.Fprint(os.Stdout, "  This token will NOT be shown again.\n")
			fmt.Fprint(os.Stdout, "  Store it safely before closing this window.\n")
			fmt.Fprint(os.Stdout, "\n")

			next := a.store.Load().Clone()
			next.Config.Security.AdminTokenHash = a.cfg.Security.AdminTokenHash
			a.store.Swap(next)
		}
	}

	// a.tray = tray.New(a.cfg, a.state)
	//
	// go a.tray.Run(func() {
	// 	a.state.SetStatus(runtime.StatusStopped)
	// 	a.Shutdown()
	// 	close(a.quitCh)
	// })

	shouldAutoStart := a.cfg.App.Mode == string(runtime.ModeAlwaysOn) && a.cfg.Startup.StartProxyOnAppLaunch
	if shouldAutoStart {
		if err := a.state.Start(); err != nil {
			log.Printf("auto-start failed: %v", err)
		}
	}

	addr := fmt.Sprintf("%s:%d", a.cfg.Server.Host, a.cfg.Server.Port)
	log.Printf("AtlasBridge v%s starting", server.Version)
	log.Printf("Listening on http://%s", addr)
	log.Printf("API endpoint: http://%s/v1", addr)
	log.Printf("Admin UI:     http://%s%s", addr, a.cfg.Server.AdminPath)
	log.Printf("Health check: http://%s/health", addr)
	log.Printf("Downstream:   %s", a.cfg.Downstream.BaseURL)
	log.Printf("Privacy mode: %s", a.cfg.Logging.PrivacyMode)
	log.Printf("Auto routing: %v", a.cfg.Routing.AutoRouting)

	if err := a.server.Serve(a.listener); err != nil && err != http.ErrServerClosed {
		a.state.SetError(err.Error())
		// a.tray.Update()
		a.errCh <- fmt.Errorf("server error: %w", err)
	}
	return nil
}

func (a *App) WaitQuit() {
	<-a.quitCh
}

func (a *App) ErrCh() <-chan error {
	return a.errCh
}

func (a *App) Shutdown() {
	log.Println("AtlasBridge shutting down...")
	a.state.SetStatus(runtime.StatusStopped)
	if a.tray != nil {
		a.tray.Update()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown timed out, forcing close: %v", err)
		a.server.Close()
	}

	startup.ReleaseLock()
	log.Println("AtlasBridge stopped.")
}
