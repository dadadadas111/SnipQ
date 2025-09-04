package clipboard

import (
	"fmt"
	"syscall"
	"unsafe"
)

// Windows API constants
const (
	CF_UNICODETEXT = 13
	GMEM_MOVEABLE  = 0x0002
)

// Windows API functions
var (
	user32               = syscall.NewLazyDLL("user32.dll")
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	procOpenClipboard    = user32.NewProc("OpenClipboard")
	procCloseClipboard   = user32.NewProc("CloseClipboard")
	procEmptyClipboard   = user32.NewProc("EmptyClipboard")
	procSetClipboardData = user32.NewProc("SetClipboardData")
	procGetClipboardData = user32.NewProc("GetClipboardData")
	procGlobalAlloc      = kernel32.NewProc("GlobalAlloc")
	procGlobalLock       = kernel32.NewProc("GlobalLock")
	procGlobalUnlock     = kernel32.NewProc("GlobalUnlock")
	procGlobalSize       = kernel32.NewProc("GlobalSize")
)

// ClipboardManager handles clipboard operations
type ClipboardManager struct {
}

// NewClipboardManager creates a new clipboard manager
func NewClipboardManager() *ClipboardManager {
	return &ClipboardManager{}
}

// SetText sets text to the clipboard
func (cm *ClipboardManager) SetText(text string) error {
	// Convert to UTF-16
	utf16Text := syscall.StringToUTF16(text)

	// Open clipboard
	ret, _, err := procOpenClipboard.Call(0)
	if ret == 0 {
		return fmt.Errorf("failed to open clipboard: %v", err)
	}
	defer procCloseClipboard.Call()

	// Empty clipboard
	ret, _, err = procEmptyClipboard.Call()
	if ret == 0 {
		return fmt.Errorf("failed to empty clipboard: %v", err)
	}

	// Allocate global memory
	dataSize := len(utf16Text) * 2 // UTF-16 is 2 bytes per character
	hMem, _, err := procGlobalAlloc.Call(GMEM_MOVEABLE, uintptr(dataSize))
	if hMem == 0 {
		return fmt.Errorf("failed to allocate memory: %v", err)
	}

	// Lock memory
	pMem, _, err := procGlobalLock.Call(hMem)
	if pMem == 0 {
		return fmt.Errorf("failed to lock memory: %v", err)
	}

	// Copy data to memory using slice header manipulation
	slice := &struct {
		data uintptr
		len  int
		cap  int
	}{
		data: pMem,
		len:  len(utf16Text),
		cap:  len(utf16Text),
	}
	destSlice := *(*[]uint16)(unsafe.Pointer(slice))
	copy(destSlice, utf16Text)

	// Unlock memory
	procGlobalUnlock.Call(hMem)

	// Set clipboard data
	ret, _, err = procSetClipboardData.Call(CF_UNICODETEXT, hMem)
	if ret == 0 {
		return fmt.Errorf("failed to set clipboard data: %v", err)
	}

	return nil
}

// GetText gets text from the clipboard
func (cm *ClipboardManager) GetText() (string, error) {
	// Open clipboard
	ret, _, err := procOpenClipboard.Call(0)
	if ret == 0 {
		return "", fmt.Errorf("failed to open clipboard: %v", err)
	}
	defer procCloseClipboard.Call()

	// Get clipboard data
	hData, _, err := procGetClipboardData.Call(CF_UNICODETEXT)
	if hData == 0 {
		return "", fmt.Errorf("failed to get clipboard data: %v", err)
	}

	// Lock memory
	pData, _, err := procGlobalLock.Call(hData)
	if pData == 0 {
		return "", fmt.Errorf("failed to lock clipboard memory: %v", err)
	}
	defer procGlobalUnlock.Call(hData)

	// Get data size
	size, _, _ := procGlobalSize.Call(hData)

	// Create slice from memory using slice header
	slice := &struct {
		data uintptr
		len  int
		cap  int
	}{
		data: pData,
		len:  int(size / 2),
		cap:  int(size / 2),
	}
	utf16Slice := *(*[]uint16)(unsafe.Pointer(slice))

	// Find null terminator
	for i, char := range utf16Slice {
		if char == 0 {
			utf16Slice = utf16Slice[:i]
			break
		}
	}

	return syscall.UTF16ToString(utf16Slice), nil
}
