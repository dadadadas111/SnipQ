package suggestions

import (
	"log"
	"syscall"
	"unsafe"
)

// Windows API constants for popup window
const (
	WS_POPUP         = 0x80000000
	WS_VISIBLE       = 0x10000000
	WS_BORDER        = 0x00800000
	WS_CAPTION       = 0x00C00000
	WS_EX_TOPMOST    = 0x00000008
	WS_EX_TOOLWINDOW = 0x00000080
	WS_EX_NOACTIVATE = 0x08000000

	SW_HIDE           = 0
	SW_SHOW           = 5
	SW_SHOWNOACTIVATE = 4

	// Messages
	WM_CREATE      = 0x0001
	WM_DESTROY     = 0x0002
	WM_PAINT       = 0x000F
	WM_ERASEBKGND  = 0x0014
	WM_LBUTTONDOWN = 0x0201
	WM_KEYDOWN     = 0x0100

	// Virtual keys
	VK_UP     = 0x26
	VK_DOWN   = 0x28
	VK_RETURN = 0x0D
	VK_ESCAPE = 0x1B

	// Colors
	COLOR_WINDOW        = 5
	COLOR_WINDOWTEXT    = 8
	COLOR_HIGHLIGHT     = 13
	COLOR_HIGHLIGHTTEXT = 14
)

// Windows API functions
var (
	user32                  = syscall.NewLazyDLL("user32.dll")
	gdi32                   = syscall.NewLazyDLL("gdi32.dll")
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	procRegisterClass       = user32.NewProc("RegisterClassW")
	procCreateWindowEx      = user32.NewProc("CreateWindowExW")
	procDestroyWindow       = user32.NewProc("DestroyWindow")
	procShowWindow          = user32.NewProc("ShowWindow")
	procSetWindowPos        = user32.NewProc("SetWindowPos")
	procGetWindowRect       = user32.NewProc("GetWindowRect")
	procGetClientRect       = user32.NewProc("GetClientRect")
	procGetDC               = user32.NewProc("GetDC")
	procReleaseDC           = user32.NewProc("ReleaseDC")
	procDefWindowProc       = user32.NewProc("DefWindowProcW")
	procGetCursorPos        = user32.NewProc("GetCursorPos")
	procGetCaretPos         = user32.NewProc("GetCaretPos")
	procGetGUIThreadInfo    = user32.NewProc("GetGUIThreadInfo")
	procGetForegroundWindow = user32.NewProc("GetForegroundWindow")
	procInvalidateRect      = user32.NewProc("InvalidateRect")
	procTextOut             = gdi32.NewProc("TextOutW")
	procGetSysColor         = user32.NewProc("GetSysColor")
	procCreateSolidBrush    = gdi32.NewProc("CreateSolidBrush")
	procSelectObject        = gdi32.NewProc("SelectObject")
	procDeleteObject        = gdi32.NewProc("DeleteObject")
	procSetBkMode           = gdi32.NewProc("SetBkMode")
	procSetTextColor        = gdi32.NewProc("SetTextColor")
	procFillRect            = user32.NewProc("FillRect")
	procDrawText            = user32.NewProc("DrawTextW")
	procGetModuleHandle     = kernel32.NewProc("GetModuleHandleW")
	procBeginPaint          = user32.NewProc("BeginPaint")
	procEndPaint            = user32.NewProc("EndPaint")
)

// WNDCLASS structure for window class registration
type WNDCLASS struct {
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
}

// PAINTSTRUCT for painting
type PAINTSTRUCT struct {
	Hdc         uintptr
	FErase      int32
	RcPaint     RECT
	FRestore    int32
	FIncUpdate  int32
	RgbReserved [32]byte
}

// Global instance for window procedure callback
var globalPopupInstance *SuggestionPopup

// Window procedure callback
func windowProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_PAINT:
		if globalPopupInstance != nil && globalPopupInstance.hwnd == hwnd {
			globalPopupInstance.handlePaint()
		}
		return 0
	case WM_DESTROY:
		return 0
	default:
		ret, _, _ := procDefWindowProc.Call(hwnd, uintptr(msg), wParam, lParam)
		return ret
	}
}

// POINT represents a point
type POINT struct {
	X, Y int32
}

// RECT represents a rectangle
type RECT struct {
	Left, Top, Right, Bottom int32
}

// GUITHREADINFO represents GUI thread information
type GUITHREADINFO struct {
	Size       uint32
	Flags      uint32
	ActiveWnd  uintptr
	FocusWnd   uintptr
	CaptureWnd uintptr
	MenuWnd    uintptr
	MoveWnd    uintptr
	CaretWnd   uintptr
	CaretRect  RECT
}

// SuggestionPopup manages the native Windows popup window
type SuggestionPopup struct {
	hwnd             uintptr
	className        *uint16
	suggestions      []SuggestionItem
	selectedIndex    int
	itemHeight       int
	windowWidth      int
	windowHeight     int
	visible          bool
	selectionHandler func(SuggestionItem)
}

// NewSuggestionPopup creates a new suggestion popup
func NewSuggestionPopup() *SuggestionPopup {
	sp := &SuggestionPopup{
		selectedIndex: -1,
		itemHeight:    25,
		windowWidth:   300,
		windowHeight:  250,
	}

	sp.registerWindowClass()
	sp.createWindow()

	return sp
}

// registerWindowClass registers the window class for the popup
func (sp *SuggestionPopup) registerWindowClass() {
	className := "SnipQSuggestionPopup"
	sp.className = syscall.StringToUTF16Ptr(className)

	// Get module handle
	hInstance, _, _ := procGetModuleHandle.Call(0)

	// Register window class
	wc := WNDCLASS{
		LpfnWndProc:   syscall.NewCallback(windowProc),
		HInstance:     hInstance,
		LpszClassName: sp.className,
		HbrBackground: uintptr(COLOR_WINDOW + 1),
	}

	procRegisterClass.Call(uintptr(unsafe.Pointer(&wc)))
}

// createWindow creates the popup window
func (sp *SuggestionPopup) createWindow() {
	// Set global instance for window procedure
	globalPopupInstance = sp

	// Get module handle
	hInstance, _, _ := procGetModuleHandle.Call(0)

	// Create the popup window with our custom class
	hwnd, _, err := procCreateWindowEx.Call(
		WS_EX_TOPMOST|WS_EX_TOOLWINDOW|WS_EX_NOACTIVATE,                        // Extended style
		uintptr(unsafe.Pointer(sp.className)),                                  // Our custom class name
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("SnipQ Suggestions"))), // Window name
		WS_POPUP|WS_BORDER, // Style
		0, 0,               // Position
		uintptr(sp.windowWidth), uintptr(sp.windowHeight), // Size
		0,         // Parent
		0,         // Menu
		hInstance, // Instance
		0,         // Param
	)

	if hwnd == 0 {
		log.Printf("[POPUP] Failed to create window: %v", err)
		return
	}

	sp.hwnd = hwnd
	log.Printf("[POPUP] Created popup window: %v", hwnd)
}

// Show displays the popup with suggestions
func (sp *SuggestionPopup) Show(suggestions []SuggestionItem, query string) {
	if sp.hwnd == 0 {
		log.Println("[POPUP] Window not created, cannot show")
		return
	}

	sp.suggestions = suggestions
	sp.selectedIndex = -1

	// Calculate window size based on suggestions
	sp.calculateSize()

	// Get cursor position to position popup nearby
	pos := sp.getCursorPosition()

	// Position the window near the cursor
	procSetWindowPos.Call(
		sp.hwnd,
		0,                                 // Insert after
		uintptr(pos.X), uintptr(pos.Y+20), // Position slightly below cursor
		uintptr(sp.windowWidth), uintptr(sp.windowHeight),
		0, // Flags
	)

	// Show the window
	procShowWindow.Call(sp.hwnd, SW_SHOWNOACTIVATE)
	sp.visible = true

	// Trigger repaint
	procInvalidateRect.Call(sp.hwnd, 0, 1)

	log.Printf("[POPUP] Showing popup with %d suggestions at (%d, %d)", len(suggestions), pos.X, pos.Y)
}

// Update updates the popup with new suggestions
func (sp *SuggestionPopup) Update(suggestions []SuggestionItem, query string) {
	if !sp.visible {
		sp.Show(suggestions, query)
		return
	}

	sp.suggestions = suggestions
	if sp.selectedIndex >= len(suggestions) {
		sp.selectedIndex = -1
	}

	sp.calculateSize()

	// Trigger repaint
	procInvalidateRect.Call(sp.hwnd, 0, 1)

	log.Printf("[POPUP] Updated popup with %d suggestions", len(suggestions))
}

// Hide hides the popup
func (sp *SuggestionPopup) Hide() {
	if sp.hwnd != 0 && sp.visible {
		procShowWindow.Call(sp.hwnd, SW_HIDE)
		sp.visible = false
		log.Println("[POPUP] Hidden popup")
	}
}

// SetSelection sets the selected item index
func (sp *SuggestionPopup) SetSelection(index int) {
	if index < 0 || index >= len(sp.suggestions) {
		sp.selectedIndex = -1
	} else {
		sp.selectedIndex = index
	}

	// Trigger repaint
	if sp.visible {
		procInvalidateRect.Call(sp.hwnd, 0, 1)
	}
}

// SetSelectionHandler sets the handler for selection events
func (sp *SuggestionPopup) SetSelectionHandler(handler func(SuggestionItem)) {
	sp.selectionHandler = handler
}

// IsVisible returns whether the popup is currently visible
func (sp *SuggestionPopup) IsVisible() bool {
	return sp.visible
}

// UpdateSelection updates the selected item (alias for SetSelection for compatibility)
func (sp *SuggestionPopup) UpdateSelection(index int) {
	sp.SetSelection(index)
}

// getCursorPosition gets the current cursor position
func (sp *SuggestionPopup) getCursorPosition() POINT {
	var pos POINT

	// Try to get caret position first (more accurate for text input)
	caretPos := sp.getCaretPosition()
	if caretPos.X != -1 && caretPos.Y != -1 {
		return caretPos
	}

	// Fallback to cursor position
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pos)))
	return pos
}

// getCaretPosition gets the current text caret position
func (sp *SuggestionPopup) getCaretPosition() POINT {
	var info GUITHREADINFO
	info.Size = uint32(unsafe.Sizeof(info))

	// Get current thread ID (0 = current thread)
	ret, _, _ := procGetGUIThreadInfo.Call(0, uintptr(unsafe.Pointer(&info)))
	if ret == 0 {
		return POINT{X: -1, Y: -1}
	}

	// If we have caret info, use it
	if info.CaretWnd != 0 {
		// Convert caret rect to screen coordinates
		return POINT{
			X: info.CaretRect.Left,
			Y: info.CaretRect.Bottom,
		}
	}

	return POINT{X: -1, Y: -1}
}

// calculateSize calculates the optimal window size
func (sp *SuggestionPopup) calculateSize() {
	if len(sp.suggestions) == 0 {
		sp.windowHeight = sp.itemHeight
		return
	}

	// Calculate height based on number of items
	itemCount := len(sp.suggestions)
	if itemCount > 10 {
		itemCount = 10 // Max 10 visible items
	}

	sp.windowHeight = itemCount*sp.itemHeight + 4 // +4 for borders

	// Width is fixed for now, could be dynamic based on content
	sp.windowWidth = 350
}

// Destroy destroys the popup window
func (sp *SuggestionPopup) Destroy() {
	if sp.hwnd != 0 {
		procDestroyWindow.Call(sp.hwnd)
		sp.hwnd = 0
		log.Println("[POPUP] Destroyed popup window")
	}
}

// handlePaint handles the WM_PAINT message to draw suggestions
func (sp *SuggestionPopup) handlePaint() {
	var ps PAINTSTRUCT
	hdc, _, _ := procBeginPaint.Call(sp.hwnd, uintptr(unsafe.Pointer(&ps)))
	if hdc == 0 {
		return
	}
	defer procEndPaint.Call(sp.hwnd, uintptr(unsafe.Pointer(&ps)))

	// Set text background to transparent
	procSetBkMode.Call(hdc, 1) // TRANSPARENT

	// Draw each suggestion
	for i, suggestion := range sp.suggestions {
		y := int32(i * sp.itemHeight)

		// Create text to display
		text := suggestion.Trigger + " - " + suggestion.Name
		if len(text) > 40 {
			text = text[:37] + "..."
		}
		textPtr := syscall.StringToUTF16Ptr(text)

		// Set colors based on selection
		if i == sp.selectedIndex {
			// Selected item - highlight
			procSetTextColor.Call(hdc, 0x00FFFFFF)                        // White text
			highlightBrush, _, _ := procCreateSolidBrush.Call(0x00FF6600) // Orange background
			rect := RECT{Left: 0, Top: y, Right: int32(sp.windowWidth), Bottom: y + int32(sp.itemHeight)}
			procFillRect.Call(hdc, uintptr(unsafe.Pointer(&rect)), highlightBrush)
			procDeleteObject.Call(highlightBrush)
		} else {
			// Normal item
			procSetTextColor.Call(hdc, 0x00000000) // Black text
		}

		// Draw the text
		rect := RECT{
			Left:   5,
			Top:    y + 3,
			Right:  int32(sp.windowWidth) - 5,
			Bottom: y + int32(sp.itemHeight) - 3,
		}
		procDrawText.Call(hdc, uintptr(unsafe.Pointer(textPtr)), uintptr(^uint(0)), uintptr(unsafe.Pointer(&rect)), 0x00000000)
	}
}
