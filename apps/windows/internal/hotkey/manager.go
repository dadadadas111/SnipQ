package hotkey

import (
	"fmt"
	"log"
	"runtime"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/snipq/core/pkg/types"
)

// Windows API constants for hotkey registration
const (
	MOD_ALT     = 0x0001
	MOD_CONTROL = 0x0002
	MOD_SHIFT   = 0x0004
	MOD_WIN     = 0x0008

	WM_HOTKEY = 0x0312

	// Virtual key codes
	VK_SPACE  = 0x20
	VK_RETURN = 0x0D
	VK_PERIOD = 0xBE // OEM_PERIOD
	VK_COMMA  = 0xBC // OEM_COMMA
	VK_A      = 0x41
	VK_Z      = 0x5A
	VK_0      = 0x30
	VK_9      = 0x39
)

// Windows API functions
var (
	user32               = syscall.NewLazyDLL("user32.dll")
	procRegisterHotKey   = user32.NewProc("RegisterHotKey")
	procUnregisterHotKey = user32.NewProc("UnregisterHotKey")
	procPeekMessage      = user32.NewProc("PeekMessageW")
	procTranslateMessage = user32.NewProc("TranslateMessage")
	procDispatchMessage  = user32.NewProc("DispatchMessageW")
)

// Message filtering constants
const (
	PM_REMOVE = 0x0001
	WM_QUIT   = 0x0012
)

// MSG represents a Windows message structure
type MSG struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
}

// EventHandler is called when a hotkey is triggered
type EventHandler func(action string)

// Manager handles hotkey registration and events
type Manager struct {
	mu            sync.RWMutex
	registrations map[string]*registration // action -> registration
	statuses      map[string]*types.HotkeyStatus
	eventHandler  EventHandler
	done          chan bool
	running       bool
	registerChan  chan registerRequest
	unregisterChan chan string
	responseChan  chan error
}

// registerRequest represents a hotkey registration request
type registerRequest struct {
	hotkeys map[string]*types.Hotkey
	initial bool
}

// registration tracks a registered hotkey
type registration struct {
	action string
	id     int32
	hotkey types.Hotkey
	hwnd   uintptr
}

// NewManager creates a new hotkey manager
func NewManager() *Manager {
	log.Println("[HOTKEY] Creating new hotkey manager")
	return &Manager{
		registrations:  make(map[string]*registration),
		statuses:       make(map[string]*types.HotkeyStatus),
		done:           make(chan bool),
		registerChan:   make(chan registerRequest),
		unregisterChan: make(chan string),
		responseChan:   make(chan error),
	}
}

// getWindowHandle gets a suitable window handle for hotkey registration
func (m *Manager) getWindowHandle() uintptr {
	// For now, always use NULL (current thread) to avoid cross-thread issues
	// The fix for the delay will come from ensuring the message loop runs efficiently
	log.Println("[HOTKEY] Using current thread (NULL) for hotkey registration")
	return 0
}

// SetEventHandler sets the callback for hotkey events
func (m *Manager) SetEventHandler(handler EventHandler) {
	log.Println("[HOTKEY] Setting event handler")
	m.mu.Lock()
	defer m.mu.Unlock()
	m.eventHandler = handler
}

// GetDefaultHotkeys returns the default hotkey configuration
func GetDefaultHotkeys() map[string]*types.Hotkey {
	log.Println("[HOTKEY] Creating default hotkey configuration")
	return map[string]*types.Hotkey{
		"openPalette": {
			Modifiers: []string{"Ctrl", "Alt"},
			Key:       "Space",
			Enabled:   true,
		},
		"togglePause": {
			Modifiers: []string{"Ctrl", "Alt"},
			Key:       "Period",
			Enabled:   true,
		},
		"expandNow": {
			Modifiers: []string{"Ctrl", "Alt"},
			Key:       "Enter",
			Enabled:   false, // Optional for v1
		},
	}
}

// ValidateHotkey validates a hotkey configuration
func ValidateHotkey(hotkey types.Hotkey) error {
	log.Printf("[HOTKEY] Validating hotkey: %+v", hotkey)

	// Must have at least one modifier
	if len(hotkey.Modifiers) == 0 {
		return fmt.Errorf("hotkey must have at least one modifier")
	}

	// Must have a key
	if hotkey.Key == "" {
		return fmt.Errorf("hotkey must have a key")
	}

	// Check for disallowed combinations (Windows reserved)
	if isReservedCombination(hotkey) {
		return fmt.Errorf("hotkey combination is reserved by Windows")
	}

	// Validate key exists in our mapping
	if _, exists := getVirtualKeyCode(hotkey.Key); !exists {
		return fmt.Errorf("unsupported key: %s", hotkey.Key)
	}

	log.Printf("[HOTKEY] Hotkey validation passed")
	return nil
}

// isReservedCombination checks if a hotkey is reserved by Windows
func isReservedCombination(hotkey types.Hotkey) bool {
	// Common Windows reserved combinations
	reserved := []struct {
		modifiers []string
		key       string
	}{
		{[]string{"Win"}, "R"},                // Run dialog
		{[]string{"Win"}, "E"},                // Explorer
		{[]string{"Win"}, "D"},                // Show desktop
		{[]string{"Alt"}, "Tab"},              // Task switcher
		{[]string{"Ctrl", "Alt"}, "Delete"},   // Security screen
		{[]string{"Ctrl", "Shift"}, "Escape"}, // Task manager
	}

	for _, res := range reserved {
		if len(res.modifiers) == len(hotkey.Modifiers) && res.key == hotkey.Key {
			// Check if modifiers match (order doesn't matter)
			modifierMap := make(map[string]bool)
			for _, mod := range hotkey.Modifiers {
				modifierMap[mod] = true
			}
			match := true
			for _, resMod := range res.modifiers {
				if !modifierMap[resMod] {
					match = false
					break
				}
			}
			if match {
				return true
			}
		}
	}
	return false
}

// getVirtualKeyCode maps key names to Windows virtual key codes
func getVirtualKeyCode(key string) (uint32, bool) {
	keyMap := map[string]uint32{
		"Space":  VK_SPACE,
		"Enter":  VK_RETURN,
		"Period": VK_PERIOD,
		"Comma":  VK_COMMA,
	}

	// Add letters A-Z
	for i := 'A'; i <= 'Z'; i++ {
		keyMap[string(i)] = uint32(VK_A + (i - 'A'))
	}

	// Add numbers 0-9
	for i := '0'; i <= '9'; i++ {
		keyMap[string(i)] = uint32(VK_0 + (i - '0'))
	}

	vk, exists := keyMap[key]
	return vk, exists
}

// getModifierFlags converts modifier names to Windows flags
func getModifierFlags(modifiers []string) uint32 {
	var flags uint32
	for _, mod := range modifiers {
		switch mod {
		case "Ctrl":
			flags |= MOD_CONTROL
		case "Alt":
			flags |= MOD_ALT
		case "Shift":
			flags |= MOD_SHIFT
		case "Win":
			flags |= MOD_WIN
		}
	}
	return flags
}

// RegisterHotkeys registers hotkeys from settings
func (m *Manager) RegisterHotkeys(hotkeys map[string]*types.Hotkey) error {
	log.Printf("[HOTKEY] Registering %d hotkeys", len(hotkeys))
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.registerHotkeysInternal(hotkeys)
}

// InitialRegisterHotkeys performs initial registration during startup (clears all first)
func (m *Manager) InitialRegisterHotkeys(hotkeys map[string]*types.Hotkey) error {
	log.Printf("[HOTKEY] Initial registration of %d hotkeys", len(hotkeys))
	
	// Check if message loop is running
	m.mu.RLock()
	running := m.running
	m.mu.RUnlock()
	
	if !running {
		log.Println("[HOTKEY] Message loop not running, starting it first")
		if err := m.StartListening(); err != nil {
			return fmt.Errorf("failed to start message loop: %v", err)
		}
		// Give the message loop time to start
		time.Sleep(100 * time.Millisecond)
	}
	
	// Send registration request to message loop thread
	select {
	case m.registerChan <- registerRequest{hotkeys: hotkeys, initial: true}:
		// Wait for response
		select {
		case err := <-m.responseChan:
			return err
		case <-time.After(5 * time.Second):
			return fmt.Errorf("timeout waiting for hotkey registration")
		}
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout sending registration request")
	}
}

// registerAllHotkeysInternal registers all hotkeys without comparing to existing (called with lock held)
func (m *Manager) registerAllHotkeysInternal(hotkeys map[string]*types.Hotkey) error {
	var errors []string
	id := int32(1)

	for action, hotkey := range hotkeys {
		if hotkey == nil || !hotkey.Enabled {
			log.Printf("[HOTKEY] Skipping disabled hotkey for action: %s", action)
			m.statuses[action] = &types.HotkeyStatus{
				Action: action,
				Status: "Not Registered",
				Error:  "Disabled",
				Hotkey: types.Hotkey{},
			}
			continue
		}

		log.Printf("[HOTKEY] Registering hotkey for action '%s': %+v", action, *hotkey)

		// Validate hotkey
		if err := ValidateHotkey(*hotkey); err != nil {
			log.Printf("[HOTKEY] Validation failed for action '%s': %v", action, err)
			m.statuses[action] = &types.HotkeyStatus{
				Action: action,
				Status: "Not Registered",
				Error:  err.Error(),
				Hotkey: *hotkey,
			}
			errors = append(errors, fmt.Sprintf("%s: %v", action, err))
			continue
		}

		// Get virtual key code and modifier flags
		vk, exists := getVirtualKeyCode(hotkey.Key)
		if !exists {
			err := fmt.Errorf("unsupported key: %s", hotkey.Key)
			log.Printf("[HOTKEY] %v", err)
			m.statuses[action] = &types.HotkeyStatus{
				Action: action,
				Status: "Not Registered",
				Error:  err.Error(),
				Hotkey: *hotkey,
			}
			errors = append(errors, fmt.Sprintf("%s: %v", action, err))
			continue
		}

		modFlags := getModifierFlags(hotkey.Modifiers)

		// Get window handle for reliable hotkey delivery
		hwnd := m.getWindowHandle()

		// Try to register with Windows
		ret, _, errno := procRegisterHotKey.Call(
			hwnd,              // hWnd - use actual window handle instead of NULL
			uintptr(id),       // id
			uintptr(modFlags), // fsModifiers
			uintptr(vk),       // vk
		)

		if ret == 0 {
			// Registration failed
			var errMsg string
			errCode := int(errno.(syscall.Errno))
			if errCode == 1409 { // ERROR_HOTKEY_ALREADY_REGISTERED
				errMsg = "Hotkey already registered by another application"
				log.Printf("[HOTKEY] Conflict for action '%s': %s", action, errMsg)
				m.statuses[action] = &types.HotkeyStatus{
					Action: action,
					Status: "Conflict",
					Error:  errMsg,
					Hotkey: *hotkey,
				}
			} else {
				errMsg = fmt.Sprintf("Registration failed with error code: %d", errCode)
				log.Printf("[HOTKEY] Error for action '%s': %s", action, errMsg)
				m.statuses[action] = &types.HotkeyStatus{
					Action: action,
					Status: "Not Registered",
					Error:  errMsg,
					Hotkey: *hotkey,
				}
			}
			errors = append(errors, fmt.Sprintf("%s: %s", action, errMsg))
			continue
		}

		// Registration successful
		log.Printf("[HOTKEY] Successfully registered hotkey for action '%s' with ID %d", action, id)
		m.registrations[action] = &registration{
			action: action,
			id:     id,
			hotkey: *hotkey,
			hwnd:   0,
		}
		m.statuses[action] = &types.HotkeyStatus{
			Action: action,
			Status: "Registered",
			Error:  "",
			Hotkey: *hotkey,
		}

		id++
	}

	if len(errors) > 0 {
		return fmt.Errorf("some hotkeys failed to register: %v", errors)
	}

	log.Printf("[HOTKEY] Initial registration completed successfully")
	return nil
}

// registerHotkeysInternal handles the actual registration logic (called with lock held)
func (m *Manager) registerHotkeysInternal(hotkeys map[string]*types.Hotkey) error {
	// First, determine which hotkeys need to be changed
	toUnregister := make([]string, 0)
	toRegister := make(map[string]*types.Hotkey)

	// Check existing registrations
	for action, existingReg := range m.registrations {
		newHotkey, exists := hotkeys[action]

		if !exists || newHotkey == nil || !newHotkey.Enabled {
			// Hotkey was removed or disabled
			log.Printf("[HOTKEY] Hotkey for action '%s' will be unregistered (removed or disabled)", action)
			toUnregister = append(toUnregister, action)
		} else if !hotkeysEqual(existingReg.hotkey, *newHotkey) {
			// Hotkey changed
			log.Printf("[HOTKEY] Hotkey for action '%s' changed from %+v to %+v", action, existingReg.hotkey, *newHotkey)
			toUnregister = append(toUnregister, action)
			toRegister[action] = newHotkey
		} else {
			// Hotkey unchanged, keep existing registration
			log.Printf("[HOTKEY] Hotkey for action '%s' unchanged, keeping existing registration", action)
		}
	}

	// Check for new hotkeys
	for action, hotkey := range hotkeys {
		if hotkey != nil && hotkey.Enabled {
			if _, exists := m.registrations[action]; !exists {
				// New hotkey
				log.Printf("[HOTKEY] New hotkey for action '%s': %+v", action, *hotkey)
				toRegister[action] = hotkey
			}
		}
	}

	// Unregister changed/removed hotkeys
	for _, action := range toUnregister {
		m.unregisterSingleInternal(action)
	}

	// Register new/changed hotkeys
	var errors []string
	id := m.getNextAvailableID()

	for action, hotkey := range toRegister {
		log.Printf("[HOTKEY] Registering hotkey for action '%s': %+v", action, *hotkey)

		// Validate hotkey
		if err := ValidateHotkey(*hotkey); err != nil {
			log.Printf("[HOTKEY] Validation failed for action '%s': %v", action, err)
			m.statuses[action] = &types.HotkeyStatus{
				Action: action,
				Status: "Not Registered",
				Error:  err.Error(),
				Hotkey: *hotkey,
			}
			errors = append(errors, fmt.Sprintf("%s: %v", action, err))
			continue
		}

		// Get virtual key code and modifier flags
		vk, exists := getVirtualKeyCode(hotkey.Key)
		if !exists {
			err := fmt.Errorf("unsupported key: %s", hotkey.Key)
			log.Printf("[HOTKEY] %v", err)
			m.statuses[action] = &types.HotkeyStatus{
				Action: action,
				Status: "Not Registered",
				Error:  err.Error(),
				Hotkey: *hotkey,
			}
			errors = append(errors, fmt.Sprintf("%s: %v", action, err))
			continue
		}

		modFlags := getModifierFlags(hotkey.Modifiers)

		// Get window handle for reliable hotkey delivery
		hwnd := m.getWindowHandle()

		// Try to register with Windows
		ret, _, errno := procRegisterHotKey.Call(
			hwnd,              // hWnd - use actual window handle instead of NULL
			uintptr(id),       // id
			uintptr(modFlags), // fsModifiers
			uintptr(vk),       // vk
		)

		if ret == 0 {
			// Registration failed
			var errMsg string
			errCode := int(errno.(syscall.Errno))
			if errCode == 1409 { // ERROR_HOTKEY_ALREADY_REGISTERED
				errMsg = "Hotkey already registered by another application"
				log.Printf("[HOTKEY] Conflict for action '%s': %s", action, errMsg)
				m.statuses[action] = &types.HotkeyStatus{
					Action: action,
					Status: "Conflict",
					Error:  errMsg,
					Hotkey: *hotkey,
				}
			} else {
				errMsg = fmt.Sprintf("Registration failed with error code: %d", errCode)
				log.Printf("[HOTKEY] Error for action '%s': %s", action, errMsg)
				m.statuses[action] = &types.HotkeyStatus{
					Action: action,
					Status: "Not Registered",
					Error:  errMsg,
					Hotkey: *hotkey,
				}
			}
			errors = append(errors, fmt.Sprintf("%s: %s", action, errMsg))
			continue
		}

		// Registration successful
		log.Printf("[HOTKEY] Successfully registered hotkey for action '%s' with ID %d", action, id)
		m.registrations[action] = &registration{
			action: action,
			id:     id,
			hotkey: *hotkey,
			hwnd:   0,
		}
		m.statuses[action] = &types.HotkeyStatus{
			Action: action,
			Status: "Registered",
			Error:  "",
			Hotkey: *hotkey,
		}

		id++
	}

	// Update status for disabled hotkeys
	for action, hotkey := range hotkeys {
		if hotkey == nil || !hotkey.Enabled {
			m.statuses[action] = &types.HotkeyStatus{
				Action: action,
				Status: "Not Registered",
				Error:  "Disabled",
				Hotkey: types.Hotkey{},
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("some hotkeys failed to register: %v", errors)
	}

	log.Printf("[HOTKEY] Registration completed successfully")
	return nil
}

// hotkeysEqual compares two hotkeys for equality
func hotkeysEqual(a, b types.Hotkey) bool {
	if a.Enabled != b.Enabled || a.Key != b.Key {
		return false
	}

	if len(a.Modifiers) != len(b.Modifiers) {
		return false
	}

	// Create maps for comparison (order doesn't matter)
	aModMap := make(map[string]bool)
	for _, mod := range a.Modifiers {
		aModMap[mod] = true
	}

	for _, mod := range b.Modifiers {
		if !aModMap[mod] {
			return false
		}
	}

	return true
}

// getNextAvailableID finds the next available hotkey ID
func (m *Manager) getNextAvailableID() int32 {
	usedIDs := make(map[int32]bool)
	for _, reg := range m.registrations {
		usedIDs[reg.id] = true
	}

	id := int32(1)
	for usedIDs[id] {
		id++
	}

	return id
}

// unregisterSingleInternal unregisters a single hotkey (called with lock held)
func (m *Manager) unregisterSingleInternal(action string) {
	if reg, exists := m.registrations[action]; exists {
		log.Printf("[HOTKEY] Unregistering hotkey for action '%s' with ID %d", action, reg.id)
		hwnd := m.getWindowHandle()
		procUnregisterHotKey.Call(hwnd, uintptr(reg.id))
		delete(m.registrations, action)
	}
}

// UnregisterAll unregisters all hotkeys
func (m *Manager) UnregisterAll() {
	log.Println("[HOTKEY] Unregistering all hotkeys")
	m.mu.Lock()
	defer m.mu.Unlock()
	m.unregisterAllInternal()
}

// unregisterAllInternal unregisters all hotkeys (called with lock held)
func (m *Manager) unregisterAllInternal() {
	hwnd := m.getWindowHandle()
	for action, reg := range m.registrations {
		log.Printf("[HOTKEY] Unregistering hotkey for action '%s' with ID %d", action, reg.id)
		procUnregisterHotKey.Call(hwnd, uintptr(reg.id))
		delete(m.registrations, action)
	}

	// Clear all statuses
	for action := range m.statuses {
		m.statuses[action] = &types.HotkeyStatus{
			Action: action,
			Status: "Not Registered",
			Error:  "",
			Hotkey: types.Hotkey{},
		}
	}
}

// GetStatuses returns the current registration status of all hotkeys
func (m *Manager) GetStatuses() map[string]*types.HotkeyStatus {
	log.Println("[HOTKEY] Getting hotkey statuses")
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Make a copy to avoid concurrent access issues
	result := make(map[string]*types.HotkeyStatus)
	for action, status := range m.statuses {
		statusCopy := *status
		result[action] = &statusCopy
	}

	log.Printf("[HOTKEY] Returning %d hotkey statuses", len(result))
	return result
}

// StartListening starts the hotkey message loop
func (m *Manager) StartListening() error {
	log.Println("[HOTKEY] Starting hotkey message loop")
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return fmt.Errorf("already listening")
	}
	m.running = true
	m.mu.Unlock()

	go m.messageLoop()
	return nil
}

// StopListening stops the hotkey message loop
func (m *Manager) StopListening() {
	log.Println("[HOTKEY] Stopping hotkey message loop")
	m.mu.Lock()
	if !m.running {
		m.mu.Unlock()
		return
	}
	m.running = false
	m.mu.Unlock()

	// Send stop signal
	select {
	case m.done <- true:
		log.Println("[HOTKEY] Stop signal sent")
	default:
		log.Println("[HOTKEY] Stop signal channel full")
	}
}

// messageLoop handles Windows messages for hotkey events
func (m *Manager) messageLoop() {
	// Lock this goroutine to an OS thread to ensure consistent message handling
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	
	log.Println("[HOTKEY] Message loop started with health monitoring on dedicated OS thread")
	defer log.Println("[HOTKEY] Message loop stopped")

	var msg MSG
	ticker := time.NewTicker(10 * time.Millisecond) // More frequent checking - 10ms instead of 50ms
	defer ticker.Stop()

	healthTicker := time.NewTicker(5 * time.Second) // Health check every 5 seconds
	defer healthTicker.Stop()

	lastMessageTime := time.Now()
	messageCount := 0

	for {
		select {
		case <-m.done:
			log.Println("[HOTKEY] Message loop received stop signal")
			return
		
		case req := <-m.registerChan:
			log.Println("[HOTKEY] Processing hotkey registration request on message loop thread")
			var err error
			if req.initial {
				// Clear all previous registrations first
				m.unregisterAllInternal()
				err = m.registerAllHotkeysInternal(req.hotkeys)
			} else {
				err = m.registerHotkeysInternal(req.hotkeys)
			}
			// Send response back
			select {
			case m.responseChan <- err:
			case <-time.After(1 * time.Second):
				log.Println("[HOTKEY] Warning: timeout sending registration response")
			}
		
		case action := <-m.unregisterChan:
			log.Printf("[HOTKEY] Processing unregister request for action: %s", action)
			m.unregisterSingleInternal(action)
			
		case <-healthTicker.C:
			// Health check - log status
			log.Printf("[HOTKEY] Message loop health check - processed %d messages, last activity: %v ago",
				messageCount, time.Since(lastMessageTime))
		case <-ticker.C:
			// Process multiple messages per cycle to handle bursts
			processedThisCycle := 0
			for processedThisCycle < 10 { // Process up to 10 messages per cycle
				// Use PeekMessage with PM_REMOVE to check for messages without blocking
				ret, _, err := procPeekMessage.Call(
					uintptr(unsafe.Pointer(&msg)),
					0,         // hWnd (any window)
					0,         // wMsgFilterMin
					0,         // wMsgFilterMax
					PM_REMOVE, // Remove message from queue
				)

				if err != nil && err.Error() != "The operation completed successfully." {
					log.Printf("[HOTKEY] PeekMessage error: %v", err)
					break
				}

				if ret == 0 { // No message available
					break
				}

				// Message available
				messageCount++
				processedThisCycle++
				lastMessageTime = time.Now()

				if msg.Message == WM_QUIT {
					log.Println("[HOTKEY] Received WM_QUIT message")
					return
				}

				// Check if it's a hotkey message
				if msg.Message == WM_HOTKEY {
					log.Printf("[HOTKEY] Processing hotkey message %d (total: %d)", msg.WParam, messageCount)
					m.handleHotkeyMessage(int32(msg.WParam))
				}

				// Translate and dispatch the message
				procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
				procDispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
			}

			if processedThisCycle > 0 {
				log.Printf("[HOTKEY] Processed %d messages in this cycle", processedThisCycle)
			}
		}
	}
} // handleHotkeyMessage handles a hotkey activation
func (m *Manager) handleHotkeyMessage(id int32) {
	log.Printf("[HOTKEY] Received hotkey message for ID: %d", id)

	// Add panic recovery for safety
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[HOTKEY] Recovered from panic in handleHotkeyMessage: %v", r)
		}
	}()

	m.mu.RLock()
	var action string
	for act, reg := range m.registrations {
		if reg.id == id {
			action = act
			break
		}
	}
	handler := m.eventHandler
	m.mu.RUnlock()

	if action == "" {
		log.Printf("[HOTKEY] No action found for hotkey ID: %d", id)
		return
	}

	log.Printf("[HOTKEY] Triggering action: %s", action)

	if handler != nil {
		// Call handler in goroutine to avoid blocking message loop
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[HOTKEY] Recovered from panic in event handler for action %s: %v", action, r)
				}
			}()
			handler(action)
		}()
	} else {
		log.Printf("[HOTKEY] No event handler set for action: %s", action)
	}
}
