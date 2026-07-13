package tray

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"github.com/atlasbridge/atlasbridge/internal/config"
	appruntime "github.com/atlasbridge/atlasbridge/internal/runtime"
	"github.com/getlantern/systray"
)

type Tray struct {
	mu          sync.Mutex
	cfg         *config.Config
	state       *appruntime.State
	proxyAddr   string
	dashURL     string
	nineDashURL string
	iconData    []byte

	titleItem    *systray.MenuItem
	statusItem   *systray.MenuItem
	endpointItem *systray.MenuItem
	openDashItem *systray.MenuItem
	startItem    *systray.MenuItem
	stopItem     *systray.MenuItem
	restartItem  *systray.MenuItem
	runAtStartup *systray.MenuItem
	alwaysOnMode *systray.MenuItem
	manualMode   *systray.MenuItem
	disabledMode *systray.MenuItem
	openNineDash *systray.MenuItem
	openLogsItem *systray.MenuItem
	copyEndpoint *systray.MenuItem
	quitItem     *systray.MenuItem
}

func New(cfg *config.Config, state *appruntime.State) *Tray {
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	dashURL := fmt.Sprintf("http://%s%s", addr, cfg.Server.AdminPath)
	nineDashURL := "http://127.0.0.1:20128/dashboard"

	iconData, _ := loadIcon()

	return &Tray{
		cfg:         cfg,
		state:       state,
		proxyAddr:   addr,
		dashURL:     dashURL,
		nineDashURL: nineDashURL,
		iconData:    iconData,
	}
}

func (t *Tray) Run(onQuit func()) {
	systray.Run(
		func() { t.onReady() },
		func() { t.onExit() },
	)
	if onQuit != nil {
		onQuit()
	}
}

func (t *Tray) onReady() {
	systray.SetIcon(t.iconData)
	systray.SetTitle("AtlasBridge")
	systray.SetTooltip("AtlasBridge - Stopped")

	t.titleItem = systray.AddMenuItem("AtlasBridge", "AtlasBridge control panel")
	t.titleItem.Disable()

	t.statusItem = systray.AddMenuItem("Status: Stopped", "Current proxy status")
	t.statusItem.Disable()

	t.endpointItem = systray.AddMenuItem("Endpoint: "+t.proxyAddr, "API endpoint address")
	t.endpointItem.Disable()

	systray.AddSeparator()

	t.openDashItem = systray.AddMenuItem("Open Dashboard", "Open admin dashboard in browser")

	systray.AddSeparator()

	t.startItem = systray.AddMenuItem("Start Proxy", "Start the proxy engine")
	t.stopItem = systray.AddMenuItem("Stop Proxy", "Stop the proxy engine")
	t.restartItem = systray.AddMenuItem("Restart Proxy", "Restart the proxy engine")

	systray.AddSeparator()

	t.runAtStartup = systray.AddMenuItemCheckbox("Run at Startup", "Toggle run at login", t.cfg.Startup.RunAtLogin)

	modeMenu := systray.AddMenuItem("Always On Mode", "Select runtime mode")
	t.alwaysOnMode = modeMenu.AddSubMenuItem("Always On", "Proxy starts automatically")
	t.manualMode = modeMenu.AddSubMenuItem("Manual", "Proxy starts manually")
	t.disabledMode = modeMenu.AddSubMenuItem("Disabled", "Proxy does not accept requests")

	systray.AddSeparator()

	t.openNineDash = systray.AddMenuItem("Open 9Router Dashboard", "Open 9Router dashboard in browser")
	t.openLogsItem = systray.AddMenuItem("Open Logs Folder", "Open log directory in file explorer")
	t.copyEndpoint = systray.AddMenuItem("Copy API Endpoint", "Copy endpoint to clipboard")

	systray.AddSeparator()

	t.quitItem = systray.AddMenuItem("Quit", "Quit AtlasBridge")

	t.updateMenuState()

	go t.handleMenuEvents()
}

func (t *Tray) onExit() {
	log.Println("tray icon removed")
}

func (t *Tray) handleMenuEvents() {
	for {
		select {
		case <-t.openDashItem.ClickedCh:
			openBrowser(t.dashURL)

		case <-t.startItem.ClickedCh:
			if err := t.state.Start(); err != nil {
				log.Printf("tray start error: %v", err)
			} else {
				log.Println("proxy started from tray")
			}

		case <-t.stopItem.ClickedCh:
			if err := t.state.Stop(); err != nil {
				log.Printf("tray stop error: %v", err)
			} else {
				log.Println("proxy stopped from tray")
			}

		case <-t.restartItem.ClickedCh:
			if err := t.state.Restart(); err != nil {
				log.Printf("tray restart error: %v", err)
			} else {
				log.Println("proxy restarted from tray")
			}

		case <-t.runAtStartup.ClickedCh:
			t.cfg.Startup.RunAtLogin = !t.cfg.Startup.RunAtLogin
			if err := config.Save(t.cfg); err != nil {
				log.Printf("tray save run_at_login error: %v", err)
				t.cfg.Startup.RunAtLogin = !t.cfg.Startup.RunAtLogin
			}

		case <-t.alwaysOnMode.ClickedCh:
			t.state.SetMode(appruntime.ModeAlwaysOn)
			t.cfg.App.Mode = string(appruntime.ModeAlwaysOn)
			if err := config.Save(t.cfg); err != nil {
				log.Printf("tray save mode error: %v", err)
			}

		case <-t.manualMode.ClickedCh:
			t.state.SetMode(appruntime.ModeManual)
			t.cfg.App.Mode = string(appruntime.ModeManual)
			if err := config.Save(t.cfg); err != nil {
				log.Printf("tray save mode error: %v", err)
			}

		case <-t.disabledMode.ClickedCh:
			t.state.SetMode(appruntime.ModeDisabled)
			t.state.SetStatus(appruntime.StatusDisabled)
			t.cfg.App.Mode = string(appruntime.ModeDisabled)
			if err := config.Save(t.cfg); err != nil {
				log.Printf("tray save mode error: %v", err)
			}

		case <-t.openNineDash.ClickedCh:
			openBrowser(t.nineDashURL)

		case <-t.openLogsItem.ClickedCh:
			openLogsFolder()

		case <-t.copyEndpoint.ClickedCh:
			copyToClipboard(fmt.Sprintf("http://%s/v1", t.proxyAddr))

		case <-t.quitItem.ClickedCh:
			systray.Quit()
			return
		}
	}
}

func (t *Tray) Update() {
	t.updateMenuState()
	systray.SetTooltip(t.buildTooltip())
	t.updateIcon()
}

func (t *Tray) updateMenuState() {
	t.mu.Lock()
	defer t.mu.Unlock()

	status := t.state.GetStatus()
	mode := t.state.GetMode()

	t.statusItem.SetTitle("Status: " + statusLabel(status))

	switch status {
	case appruntime.StatusRunning:
		t.startItem.Disable()
		t.stopItem.Enable()
		t.restartItem.Enable()
	case appruntime.StatusStopped:
		t.startItem.Enable()
		t.stopItem.Disable()
		t.restartItem.Disable()
	case appruntime.StatusDisabled:
		t.startItem.Disable()
		t.stopItem.Disable()
		t.restartItem.Disable()
	case appruntime.StatusError:
		t.startItem.Enable()
		t.stopItem.Disable()
		t.restartItem.Enable()
	default:
		t.startItem.Disable()
		t.stopItem.Disable()
		t.restartItem.Disable()
	}

	if t.cfg.Startup.RunAtLogin {
		t.runAtStartup.Check()
	} else {
		t.runAtStartup.Uncheck()
	}

	t.alwaysOnMode.Uncheck()
	t.manualMode.Uncheck()
	t.disabledMode.Uncheck()

	switch mode {
	case appruntime.ModeAlwaysOn:
		t.alwaysOnMode.Check()
	case appruntime.ModeManual:
		t.manualMode.Check()
	case appruntime.ModeDisabled:
		t.disabledMode.Check()
	}
}

func (t *Tray) buildTooltip() string {
	status := t.state.GetStatus()
	mode := t.state.GetMode()
	return fmt.Sprintf("AtlasBridge - %s (%s) - %s", statusLabel(status), mode, t.proxyAddr)
}

func (t *Tray) updateIcon() {
	// On Windows, dynamic icon updates via LoadImage don't work well with ICO format
	// Keep the same icon; status is shown in tooltip instead
}

func statusLabel(s appruntime.Status) string {
	switch s {
	case appruntime.StatusRunning:
		return "Running"
	case appruntime.StatusStopped:
		return "Stopped"
	case appruntime.StatusDisabled:
		return "Disabled"
	case appruntime.StatusError:
		return "Error"
	case appruntime.StatusPortConflict:
		return "Port Conflict"
	case appruntime.StatusDownstreamOff:
		return "Downstream Offline"
	case appruntime.StatusStarting:
		return "Starting"
	default:
		return string(s)
	}
}

func loadIcon() ([]byte, error) {
	if len(defaultIcon) > 0 {
		return defaultIcon, nil
	}
	return nil, nil
}

func openBrowser(urlStr string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", urlStr)
	case "darwin":
		cmd = exec.Command("open", urlStr)
	default:
		cmd = exec.Command("xdg-open", urlStr)
	}
	if err := cmd.Start(); err != nil {
		log.Printf("failed to open browser: %v", err)
	}
}

func openLogsFolder() {
	logDir := config.ConfigDir()
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", logDir)
	case "darwin":
		cmd = exec.Command("open", logDir)
	default:
		cmd = exec.Command("xdg-open", logDir)
	}
	if err := cmd.Start(); err != nil {
		log.Printf("failed to open logs folder: %v", err)
	}
}

func copyToClipboard(text string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("powershell", "-NoProfile", "-Command", "Set-Clipboard", "--", text)
	case "darwin":
		cmd = exec.Command("pbcopy")
	default:
		cmd = exec.Command("xclip", "-selection", "clipboard")
	}
	if runtime.GOOS != "windows" {
		cmd.Stdin = strings.NewReader(text)
	}
	if err := cmd.Run(); err != nil {
		log.Printf("failed to copy to clipboard: %v", err)
	}
}
