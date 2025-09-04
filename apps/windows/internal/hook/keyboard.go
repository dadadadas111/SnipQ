package hook

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/snipq/core/pkg/types"
)

// Windows API constants
const (
	WH_KEYBOARD_LL = 13
	WM_KEYDOWN     = 0x0100
	WM_SYSKEYDOWN  = 0x0104
	HC_ACTION      = 0

	// Virtual key codes
	VK_BACK    = 0x08
	VK_TAB     = 0x09
	VK_RETURN  = 0x0D
	VK_SHIFT   = 0x10
	VK_CONTROL = 0x11
	VK_ALT     = 0x12
	VK_SPACE   = 0x20
	VK_DELETE  = 0x2E
	VK_A       = 0x41
	VK_Z       = 0x5A
	VK_0       = 0x30
	VK_9       = 0x39

	// Arrow keys for suggestion navigation
	VK_UP   = 0x26
	VK_DOWN = 0x28

	// Special characters
	VK_OEM_1      = 0xBA // ';:' key
	VK_OEM_PLUS   = 0xBB // '=+' key
	VK_OEM_COMMA  = 0xBC // ',<' key
	VK_OEM_MINUS  = 0xBD // '-_' key
	VK_OEM_PERIOD = 0xBE // '.>' key
	VK_OEM_2      = 0xBF // '/?' key
	VK_OEM_3      = 0xC0 // '`~' key
	VK_OEM_4      = 0xDB // '[{' key
	VK_OEM_5      = 0xDC // '\|' key
	VK_OEM_6      = 0xDD // ']}' key
	VK_OEM_7      = 0xDE // ''"' key
)

// Windows API functions
var (
	user32                  = syscall.NewLazyDLL("user32.dll")
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	procSetWindowsHookEx    = user32.NewProc("SetWindowsHookExW")
	procUnhookWindowsHookEx = user32.NewProc("UnhookWindowsHookEx")
	procCallNextHookEx      = user32.NewProc("CallNextHookEx")
	procGetMessage          = user32.NewProc("GetMessageW")
	procTranslateMessage    = user32.NewProc("TranslateMessage")
	procDispatchMessage     = user32.NewProc("DispatchMessageW")
	procGetModuleHandle     = kernel32.NewProc("GetModuleHandleW")
	procGetKeyState         = user32.NewProc("GetKeyState")
)

// KBDLLHOOKSTRUCT represents the keyboard hook structure
type KBDLLHOOKSTRUCT struct {
	VkCode      uint32
	ScanCode    uint32
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}

// MSG represents a Windows message
type MSG struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
}

// TriggerHandler is called when a trigger is detected
type TriggerHandler func(trigger string)

// BufferUpdateHandler is called when the buffer content changes
type BufferUpdateHandler func(buffer string)

// NavigationHandler is called when navigation events occur
type NavigationHandler func(direction string) bool

// SuggestionAcceptHandler is called to check if suggestions should be accepted instead of normal expansion
type SuggestionAcceptHandler func() bool

// KeyboardHook manages the low-level keyboard hook
type KeyboardHook struct {
	mu                sync.RWMutex
	hook              uintptr
	running           bool
	buffer            strings.Builder
	lastKeyTime       time.Time
	prefix            string
	expandKey         string
	handler           TriggerHandler
	bufferHandler     BufferUpdateHandler
	navHandler        NavigationHandler
	suggestionHandler SuggestionAcceptHandler
	done              chan bool
	bufferMutex       sync.Mutex

	// Settings
	strictBoundaries bool
	maxBufferSize    int
	typingTimeout    time.Duration
}

// NewKeyboardHook creates a new keyboard hook
func NewKeyboardHook() *KeyboardHook {
	return &KeyboardHook{
		prefix:           ":",
		expandKey:        "Tab",
		maxBufferSize:    100,
		typingTimeout:    2 * time.Second,
		strictBoundaries: true,
		done:             make(chan bool),
	}
}

// SetTriggerHandler sets the handler for trigger detection
func (kh *KeyboardHook) SetTriggerHandler(handler TriggerHandler) {
	kh.mu.Lock()
	defer kh.mu.Unlock()
	kh.handler = handler
}

// SetBufferHandler sets the handler for buffer updates
func (kh *KeyboardHook) SetBufferHandler(handler BufferUpdateHandler) {
	kh.mu.Lock()
	defer kh.mu.Unlock()
	kh.bufferHandler = handler
}

// SetNavigationHandler sets the handler for navigation events
func (kh *KeyboardHook) SetNavigationHandler(handler NavigationHandler) {
	kh.mu.Lock()
	defer kh.mu.Unlock()
	kh.navHandler = handler
}

// SetSuggestionAcceptHandler sets the handler for suggestion acceptance
func (kh *KeyboardHook) SetSuggestionAcceptHandler(handler SuggestionAcceptHandler) {
	kh.mu.Lock()
	defer kh.mu.Unlock()
	kh.suggestionHandler = handler
}

// UpdateSettings updates the hook settings
func (kh *KeyboardHook) UpdateSettings(settings types.Settings) {
	kh.mu.Lock()
	defer kh.mu.Unlock()

	kh.prefix = settings.Prefix
	kh.expandKey = settings.ExpandKey
	kh.strictBoundaries = settings.StrictBoundaries

	log.Printf("[HOOK] Settings updated: prefix=%s, expandKey=%s, strictBoundaries=%t",
		kh.prefix, kh.expandKey, kh.strictBoundaries)
}

// Start starts the keyboard hook
func (kh *KeyboardHook) Start() error {
	kh.mu.Lock()
	if kh.running {
		kh.mu.Unlock()
		return fmt.Errorf("keyboard hook is already running")
	}
	kh.running = true
	kh.mu.Unlock()

	log.Println("[HOOK] Starting keyboard hook")

	// Start the hook in a separate goroutine
	go kh.runHook()

	return nil
}

// Stop stops the keyboard hook
func (kh *KeyboardHook) Stop() {
	kh.mu.Lock()
	if !kh.running {
		kh.mu.Unlock()
		return
	}
	kh.running = false
	kh.mu.Unlock()

	log.Println("[HOOK] Stopping keyboard hook")

	// Send stop signal
	select {
	case kh.done <- true:
	default:
	}
}

// IsRunning returns whether the hook is currently running
func (kh *KeyboardHook) IsRunning() bool {
	kh.mu.RLock()
	defer kh.mu.RUnlock()
	return kh.running
}

// ClearBuffer clears the internal key buffer
func (kh *KeyboardHook) ClearBuffer() {
	kh.bufferMutex.Lock()
	defer kh.bufferMutex.Unlock()
	kh.buffer.Reset()
	log.Println("[HOOK] Buffer cleared")
}

// runHook runs the keyboard hook in a dedicated thread
func (kh *KeyboardHook) runHook() {
	// Lock this goroutine to an OS thread
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	log.Println("[HOOK] Installing keyboard hook")

	// Get module handle
	hModule, _, _ := procGetModuleHandle.Call(0)

	// Install the hook
	hookProc := syscall.NewCallback(kh.keyboardProc)
	hook, _, err := procSetWindowsHookEx.Call(
		WH_KEYBOARD_LL,
		hookProc,
		hModule,
		0,
	)

	if hook == 0 {
		log.Printf("[HOOK] Failed to install keyboard hook: %v", err)
		kh.mu.Lock()
		kh.running = false
		kh.mu.Unlock()
		return
	}

	kh.hook = hook
	log.Printf("[HOOK] Keyboard hook installed successfully (handle: %v)", hook)

	// Message loop
	var msg MSG
	for {
		select {
		case <-kh.done:
			log.Println("[HOOK] Received stop signal")
			goto cleanup
		default:
		}

		// Get message with timeout
		ret, _, _ := procGetMessage.Call(
			uintptr(unsafe.Pointer(&msg)),
			0,
			0,
			0,
		)

		if ret == 0 { // WM_QUIT
			log.Println("[HOOK] Received WM_QUIT")
			break
		} else if ret == ^uintptr(0) { // Error
			log.Println("[HOOK] GetMessage error")
			break
		}

		// Process message
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
	}

cleanup:
	// Unhook
	if kh.hook != 0 {
		log.Println("[HOOK] Unhooking keyboard hook")
		procUnhookWindowsHookEx.Call(kh.hook)
		kh.hook = 0
	}

	kh.mu.Lock()
	kh.running = false
	kh.mu.Unlock()

	log.Println("[HOOK] Keyboard hook stopped")
}

// keyboardProc is the low-level keyboard hook procedure
func (kh *KeyboardHook) keyboardProc(nCode int, wParam uintptr, lParam uintptr) uintptr {
	if nCode < HC_ACTION {
		ret, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, lParam)
		return ret
	}

	// Only process key down events
	if wParam != WM_KEYDOWN && wParam != WM_SYSKEYDOWN {
		ret, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, lParam)
		return ret
	}

	// Get keyboard data - need to handle the pointer conversion carefully
	kbdData := *(*KBDLLHOOKSTRUCT)(unsafe.Pointer(lParam))
	vkCode := kbdData.VkCode

	// Process the key
	suppressKey := kh.processKey(vkCode)

	if suppressKey {
		// Suppress the key by not calling CallNextHookEx
		return 1
	}

	// Pass the key through
	ret, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, lParam)
	return ret
}

// processKey processes a key press and returns whether to suppress it
func (kh *KeyboardHook) processKey(vkCode uint32) bool {
	now := time.Now()

	kh.bufferMutex.Lock()
	defer kh.bufferMutex.Unlock()

	// Check for typing timeout
	if !kh.lastKeyTime.IsZero() && now.Sub(kh.lastKeyTime) > kh.typingTimeout {
		kh.buffer.Reset()
	}
	kh.lastKeyTime = now

	// Handle special keys
	switch vkCode {
	case VK_UP:
		// Handle up arrow for suggestion navigation
		if kh.navHandler != nil {
			// Try to handle as navigation, suppress key if handled
			if kh.navHandler("up") {
				return true // Suppress the key to prevent cursor movement
			}
		}
		return false

	case VK_DOWN:
		// Handle down arrow for suggestion navigation
		if kh.navHandler != nil {
			// Try to handle as navigation, suppress key if handled
			if kh.navHandler("down") {
				return true // Suppress the key to prevent cursor movement
			}
		}
		return false

	case VK_BACK:
		// Remove last character from buffer
		if kh.buffer.Len() > 0 {
			content := kh.buffer.String()
			if len(content) > 0 {
				// Handle UTF-8 properly
				runes := []rune(content)
				if len(runes) > 0 {
					kh.buffer.Reset()
					kh.buffer.WriteString(string(runes[:len(runes)-1]))
				}
			}
		}
		// Notify buffer handler after backspace
		kh.notifyBufferUpdate()
		return false

	case VK_RETURN, VK_SPACE:
		// Check if Enter should be handled as suggestion acceptance
		if vkCode == VK_RETURN && kh.navHandler != nil {
			// Try to handle as accept, suppress key if handled
			if kh.navHandler("accept") {
				return true // Suppress the key
			}
		}

		// These keys break trigger sequences
		if kh.strictBoundaries {
			kh.buffer.Reset()
			// Notify buffer handler after reset
			kh.notifyBufferUpdate()
		}
		return false

	case VK_TAB:
		// Check if this is the expand key
		if kh.expandKey == "Tab" {
			return kh.handleExpandKey()
		}
		return false

	case VK_DELETE, VK_SHIFT, VK_CONTROL, VK_ALT:
		// Ignore these keys
		return false
	}

	// Convert virtual key to character
	char := kh.vkToChar(vkCode)
	if char == "" {
		return false
	}

	// Add character to buffer
	kh.buffer.WriteString(char)

	// Limit buffer size
	if kh.buffer.Len() > kh.maxBufferSize {
		content := kh.buffer.String()
		kh.buffer.Reset()
		// Keep the last part of the buffer
		if len(content) > 50 {
			kh.buffer.WriteString(content[len(content)-50:])
		} else {
			kh.buffer.WriteString(content)
		}
	}

	// Notify buffer handler
	kh.notifyBufferUpdate()

	return false
}

// handleExpandKey handles the expand key press
func (kh *KeyboardHook) handleExpandKey() bool {
	// First, check if there are suggestions available to accept
	kh.mu.RLock()
	suggestionHandler := kh.suggestionHandler
	kh.mu.RUnlock()

	if suggestionHandler != nil && suggestionHandler() {
		// Suggestion was accepted, suppress the key
		return true
	}

	// No suggestions accepted, proceed with normal trigger expansion
	content := kh.buffer.String()

	// Check if we have a potential trigger
	if !strings.Contains(content, kh.prefix) {
		return false
	}

	// Find the last occurrence of the prefix
	lastPrefixIndex := strings.LastIndex(content, kh.prefix)
	if lastPrefixIndex == -1 {
		return false
	}

	// Extract the potential trigger
	trigger := content[lastPrefixIndex:]

	// Validate trigger (basic validation)
	if len(trigger) < 2 { // At least prefix + one character
		return false
	}

	// Check for word boundaries if strict mode is enabled
	if kh.strictBoundaries {
		if lastPrefixIndex > 0 {
			prevChar := content[lastPrefixIndex-1]
			if isAlphanumeric(prevChar) {
				return false
			}
		}
	}

	log.Printf("[HOOK] Detected trigger: %s", trigger)

	// Clear the buffer
	kh.buffer.Reset()

	// Call the trigger handler
	kh.mu.RLock()
	handler := kh.handler
	kh.mu.RUnlock()

	if handler != nil {
		// Call handler in a goroutine to avoid blocking the hook
		go handler(trigger)
	}

	// Suppress the Tab key since we handled the trigger
	return true
}

// vkToChar converts a virtual key code to a character
func (kh *KeyboardHook) vkToChar(vkCode uint32) string {
	// Check for shift state
	shiftPressed := kh.isKeyPressed(VK_SHIFT)

	// Handle letters A-Z
	if vkCode >= VK_A && vkCode <= VK_Z {
		char := byte('a' + (vkCode - VK_A))
		if shiftPressed {
			char = byte('A' + (vkCode - VK_A))
		}
		return string(char)
	}

	// Handle numbers 0-9
	if vkCode >= VK_0 && vkCode <= VK_9 {
		if shiftPressed {
			// Shifted number keys
			symbols := ")!@#$%^&*("
			return string(symbols[vkCode-VK_0])
		}
		return string(byte('0' + (vkCode - VK_0)))
	}

	// Handle special characters
	switch vkCode {
	case VK_SPACE:
		return " "
	case VK_OEM_1: // ';:'
		if shiftPressed {
			return ":"
		}
		return ";"
	case VK_OEM_PLUS: // '=+'
		if shiftPressed {
			return "+"
		}
		return "="
	case VK_OEM_COMMA: // ',<'
		if shiftPressed {
			return "<"
		}
		return ","
	case VK_OEM_MINUS: // '-_'
		if shiftPressed {
			return "_"
		}
		return "-"
	case VK_OEM_PERIOD: // '.>'
		if shiftPressed {
			return ">"
		}
		return "."
	case VK_OEM_2: // '/?'
		if shiftPressed {
			return "?"
		}
		return "/"
	case VK_OEM_3: // '`~'
		if shiftPressed {
			return "~"
		}
		return "`"
	case VK_OEM_4: // '[{'
		if shiftPressed {
			return "{"
		}
		return "["
	case VK_OEM_5: // '\|'
		if shiftPressed {
			return "|"
		}
		return "\\"
	case VK_OEM_6: // ']}'
		if shiftPressed {
			return "}"
		}
		return "]"
	case VK_OEM_7: // ''"'
		if shiftPressed {
			return "\""
		}
		return "'"
	}

	return ""
}

// isKeyPressed checks if a key is currently pressed
func (kh *KeyboardHook) isKeyPressed(vkCode uint32) bool {
	ret, _, _ := procGetKeyState.Call(uintptr(vkCode))
	return (ret & 0x8000) != 0
}

// isAlphanumeric checks if a byte represents an alphanumeric character
func isAlphanumeric(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}

// GetCurrentBuffer returns the current buffer content (for debugging)
func (kh *KeyboardHook) GetCurrentBuffer() string {
	kh.bufferMutex.Lock()
	defer kh.bufferMutex.Unlock()
	return kh.buffer.String()
}

// notifyBufferUpdate notifies the buffer handler of buffer changes
func (kh *KeyboardHook) notifyBufferUpdate() {
	// Get handlers
	kh.mu.RLock()
	bufferHandler := kh.bufferHandler
	kh.mu.RUnlock()

	if bufferHandler != nil {
		// Call handler in a goroutine to avoid blocking the hook
		content := kh.buffer.String()
		go bufferHandler(content)
	}
}
