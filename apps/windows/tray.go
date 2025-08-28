package main

import (
	"context"
	"log"

	"github.com/getlantern/systray"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// TrayManager handles system tray functionality
type TrayManager struct {
	app       *App
	ctx       context.Context
	isRunning bool
	menuReady bool
}

// NewTrayManager creates a new tray manager
func NewTrayManager(app *App) *TrayManager {
	return &TrayManager{
		app: app,
	}
}

// Initialize sets up the system tray
func (t *TrayManager) Initialize(ctx context.Context) error {
	t.ctx = ctx

	// Setup system tray
	err := t.setupSystemTray()
	if err != nil {
		log.Printf("Failed to setup system tray: %v", err)
		return err
	}

	log.Println("System tray initialized successfully")
	return nil
}

// setupSystemTray creates the real system tray with icon and menu
func (t *TrayManager) setupSystemTray() error {
	log.Println("Setting up real system tray with icon and menu")

	// Start systray in a goroutine to prevent blocking
	go func() {
		systray.Run(t.onSystrayReady, t.onSystrayExit)
	}()

	t.isRunning = true
	return nil
}

// onSystrayReady is called when the system tray is ready
func (t *TrayManager) onSystrayReady() {
	// Skip icon for now to focus on menu functionality
	// systray.SetIcon(iconData)
	systray.SetTitle("SnipQ")
	systray.SetTooltip("SnipQ - Universal Snippet Expander")

	// Create menu items
	mShow := systray.AddMenuItem("Show SnipQ", "Show the main window")
	mQuickExpand := systray.AddMenuItem("Quick Expand...", "Open for quick snippet expansion")
	systray.AddSeparator()
	mAbout := systray.AddMenuItem("About SnipQ", "About this application")
	systray.AddSeparator()
	mExit := systray.AddMenuItem("Exit", "Exit the application")

	t.menuReady = true
	log.Println("System tray menu initialized and ready")

	// Handle menu clicks in a persistent goroutine
	go t.handleMenuClicks(mShow, mQuickExpand, mAbout, mExit)
}

// handleMenuClicks handles all menu click events in a separate goroutine
func (t *TrayManager) handleMenuClicks(mShow, mQuickExpand, mAbout, mExit *systray.MenuItem) {
	for {
		select {
		case <-mShow.ClickedCh:
			log.Println("Tray: Show SnipQ clicked")
			t.ShowWindow()
		case <-mQuickExpand.ClickedCh:
			log.Println("Tray: Quick Expand clicked")
			t.ShowWindow()
			// Emit focus-input event
			runtime.EventsEmit(t.ctx, "focus-input")
		case <-mAbout.ClickedCh:
			log.Println("Tray: About clicked")
			t.showAbout()
		case <-mExit.ClickedCh:
			log.Println("Tray: Exit clicked")
			t.ExitApp()
			return
		}
	}
}

// onSystrayExit is called when the system tray exits
func (t *TrayManager) onSystrayExit() {
	log.Println("System tray exited")
	t.isRunning = false
}

// showAbout shows information about the app
func (t *TrayManager) showAbout() {
	runtime.MessageDialog(t.ctx, runtime.MessageDialogOptions{
		Type:          runtime.InfoDialog,
		Title:         "About SnipQ",
		Message:       "SnipQ - Universal Snippet Expander\nVersion 1.0\n\nA powerful snippet management and expansion tool.",
		DefaultButton: "OK",
	})
}

// ToggleWindow shows or hides the main application window
func (t *TrayManager) ToggleWindow() {
	log.Println("Tray: Toggling window visibility")
	if runtime.WindowIsMinimised(t.ctx) {
		log.Println("Tray: Window is minimized, showing it")
		runtime.WindowShow(t.ctx)
		runtime.WindowUnminimise(t.ctx)
	} else {
		log.Println("Tray: Hiding window to tray")
		runtime.WindowHide(t.ctx)
	}
}

// ShowWindow shows the window and brings it to front
func (t *TrayManager) ShowWindow() {
	log.Println("Tray: Showing window")
	runtime.WindowShow(t.ctx)
	runtime.WindowUnminimise(t.ctx)
	runtime.WindowSetAlwaysOnTop(t.ctx, true)
	runtime.WindowSetAlwaysOnTop(t.ctx, false) // Bring to front then reset

	// Emit event to focus on input field
	runtime.EventsEmit(t.ctx, "focus-input")
}

// HideWindow hides the window (minimize to tray effect)
func (t *TrayManager) HideWindow() {
	log.Println("Tray: Hiding window to tray")
	runtime.WindowHide(t.ctx)
}

// ExitApp completely exits the application
func (t *TrayManager) ExitApp() {
	if t.isRunning {
		systray.Quit()
	}
	runtime.Quit(t.ctx)
}
