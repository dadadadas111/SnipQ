package main

import (
	"context"
	"log"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// TrayManager handles system tray functionality
type TrayManager struct {
	app *App
	ctx context.Context
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

// setupSystemTray creates the system tray
func (t *TrayManager) setupSystemTray() error {
	// For now, we'll implement basic window management
	// System tray with full menu will be added in the next iteration
	log.Println("Setting up basic window management for tray functionality")
	return nil
}

// ToggleWindow shows or hides the main application window
func (t *TrayManager) ToggleWindow() {
	if runtime.WindowIsMinimised(t.ctx) {
		runtime.WindowShow(t.ctx)
		runtime.WindowUnminimise(t.ctx)
	} else {
		runtime.WindowHide(t.ctx)
	}
}

// ShowWindow shows the window and brings it to front
func (t *TrayManager) ShowWindow() {
	runtime.WindowShow(t.ctx)
	runtime.WindowUnminimise(t.ctx)
	runtime.WindowSetAlwaysOnTop(t.ctx, true)
	runtime.WindowSetAlwaysOnTop(t.ctx, false) // Bring to front then reset

	// Emit event to focus on input field
	runtime.EventsEmit(t.ctx, "focus-input")
}

// HideWindow hides the window (minimize to tray effect)
func (t *TrayManager) HideWindow() {
	runtime.WindowHide(t.ctx)
}

// ExitApp completely exits the application
func (t *TrayManager) ExitApp() {
	runtime.Quit(t.ctx)
}
