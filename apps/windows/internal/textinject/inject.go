package textinject

import (
	"fmt"
	"log"
	"syscall"
	"time"
	"unsafe"
)

// Windows API constants
const (
	INPUT_KEYBOARD    = 1
	KEYEVENTF_KEYUP   = 0x0002
	KEYEVENTF_UNICODE = 0x0004

	// Virtual key codes for special keys
	VK_BACK    = 0x08
	VK_CONTROL = 0x11
	VK_SHIFT   = 0x10
	VK_ALT     = 0x12
	VK_V       = 0x56
	VK_A       = 0x41
	VK_C       = 0x43
	VK_Z       = 0x5A
)

// Windows API functions
var (
	user32          = syscall.NewLazyDLL("user32.dll")
	kernel32        = syscall.NewLazyDLL("kernel32.dll")
	procSendInput   = user32.NewProc("SendInput")
	procGetKeyState = user32.NewProc("GetKeyState")
	procSleep       = kernel32.NewProc("Sleep")
)

// INPUT structure for SendInput
type INPUT struct {
	Type uint32
	Ki   KEYBDINPUT
	_    [8]byte // padding to match C struct size
}

// KEYBDINPUT structure
type KEYBDINPUT struct {
	VirtualKeyCode uint16
	ScanCode       uint16
	Flags          uint32
	Time           uint32
	ExtraInfo      uintptr
}

// TextInjector handles text injection into active applications
type TextInjector struct {
}

// NewTextInjector creates a new text injector
func NewTextInjector() *TextInjector {
	return &TextInjector{}
}

// InjectText injects text into the currently active application
func (ti *TextInjector) InjectText(text string) error {
	log.Printf("[TEXTINJECT] Injecting text: %s", text)

	// Small delay to ensure the hotkey release is processed
	time.Sleep(50 * time.Millisecond)

	// Send the text using Unicode input
	return ti.sendUnicodeText(text)
}

// ReplaceText replaces text by first deleting characters and then injecting new text
func (ti *TextInjector) ReplaceText(deleteCount int, newText string) error {
	log.Printf("[TEXTINJECT] Replacing text: delete %d chars, insert '%s'", deleteCount, newText)

	// Small delay to ensure previous operations are processed
	time.Sleep(50 * time.Millisecond)

	// Delete the specified number of characters
	if deleteCount > 0 {
		err := ti.sendBackspaces(deleteCount)
		if err != nil {
			return fmt.Errorf("failed to delete characters: %w", err)
		}
	}

	// Small delay between delete and insert
	time.Sleep(25 * time.Millisecond)

	// Insert the new text
	return ti.sendUnicodeText(newText)
}

// SendBackspaces sends the specified number of backspace key presses
func (ti *TextInjector) SendBackspaces(count int) error {
	log.Printf("[TEXTINJECT] Sending %d backspaces", count)

	for i := 0; i < count; i++ {
		// Send backspace key down
		err := ti.sendKeyEvent(VK_BACK, false)
		if err != nil {
			return fmt.Errorf("failed to send backspace down: %w", err)
		}

		// Send backspace key up
		err = ti.sendKeyEvent(VK_BACK, true)
		if err != nil {
			return fmt.Errorf("failed to send backspace up: %w", err)
		}

		// Small delay between keystrokes
		time.Sleep(10 * time.Millisecond)
	}

	return nil
}

// sendBackspaces sends the specified number of backspace key presses (internal)
func (ti *TextInjector) sendBackspaces(count int) error {
	log.Printf("[TEXTINJECT] Sending %d backspaces", count)

	for i := 0; i < count; i++ {
		// Send backspace key down
		err := ti.sendKeyEvent(VK_BACK, false)
		if err != nil {
			return fmt.Errorf("failed to send backspace down: %w", err)
		}

		// Send backspace key up
		err = ti.sendKeyEvent(VK_BACK, true)
		if err != nil {
			return fmt.Errorf("failed to send backspace up: %w", err)
		}

		// Small delay between keystrokes
		time.Sleep(10 * time.Millisecond)
	}

	return nil
}

// sendUnicodeText sends text using Unicode input
func (ti *TextInjector) sendUnicodeText(text string) error {
	log.Printf("[TEXTINJECT] Sending Unicode text: %s", text)

	// Convert string to UTF-16
	utf16Text := syscall.StringToUTF16(text)

	for _, char := range utf16Text {
		if char == 0 { // Skip null terminator
			break
		}

		// Send character down
		err := ti.sendUnicodeChar(char, false)
		if err != nil {
			return fmt.Errorf("failed to send character down: %w", err)
		}

		// Send character up
		err = ti.sendUnicodeChar(char, true)
		if err != nil {
			return fmt.Errorf("failed to send character up: %w", err)
		}

		// Small delay between characters for better compatibility
		time.Sleep(5 * time.Millisecond)
	}

	return nil
}

// sendUnicodeChar sends a single Unicode character
func (ti *TextInjector) sendUnicodeChar(char uint16, keyUp bool) error {
	var flags uint32 = KEYEVENTF_UNICODE
	if keyUp {
		flags |= KEYEVENTF_KEYUP
	}

	input := INPUT{
		Type: INPUT_KEYBOARD,
		Ki: KEYBDINPUT{
			VirtualKeyCode: 0,
			ScanCode:       char,
			Flags:          flags,
			Time:           0,
			ExtraInfo:      0,
		},
	}

	ret, _, err := procSendInput.Call(
		1,
		uintptr(unsafe.Pointer(&input)),
		unsafe.Sizeof(input),
	)

	if ret == 0 {
		return fmt.Errorf("SendInput failed: %v", err)
	}

	return nil
}

// sendKeyEvent sends a virtual key event
func (ti *TextInjector) sendKeyEvent(vkCode uint16, keyUp bool) error {
	var flags uint32 = 0
	if keyUp {
		flags |= KEYEVENTF_KEYUP
	}

	input := INPUT{
		Type: INPUT_KEYBOARD,
		Ki: KEYBDINPUT{
			VirtualKeyCode: vkCode,
			ScanCode:       0,
			Flags:          flags,
			Time:           0,
			ExtraInfo:      0,
		},
	}

	ret, _, err := procSendInput.Call(
		1,
		uintptr(unsafe.Pointer(&input)),
		unsafe.Sizeof(input),
	)

	if ret == 0 {
		return fmt.Errorf("SendInput failed: %v", err)
	}

	return nil
}

// IsModifierPressed checks if a modifier key is currently pressed
func (ti *TextInjector) IsModifierPressed(vkCode uint16) bool {
	ret, _, _ := procGetKeyState.Call(uintptr(vkCode))
	return (ret & 0x8000) != 0
}

// SendClipboardPaste sends Ctrl+V to paste from clipboard (alternative method)
func (ti *TextInjector) SendClipboardPaste() error {
	log.Printf("[TEXTINJECT] Sending Ctrl+V")

	// Press Ctrl
	err := ti.sendKeyEvent(VK_CONTROL, false)
	if err != nil {
		return fmt.Errorf("failed to press Ctrl: %w", err)
	}

	// Small delay
	time.Sleep(10 * time.Millisecond)

	// Press V
	err = ti.sendKeyEvent(VK_V, false)
	if err != nil {
		// Release Ctrl before returning error
		ti.sendKeyEvent(VK_CONTROL, true)
		return fmt.Errorf("failed to press V: %w", err)
	}

	// Small delay
	time.Sleep(10 * time.Millisecond)

	// Release V
	err = ti.sendKeyEvent(VK_V, true)
	if err != nil {
		// Release Ctrl before returning error
		ti.sendKeyEvent(VK_CONTROL, true)
		return fmt.Errorf("failed to release V: %w", err)
	}

	// Small delay
	time.Sleep(10 * time.Millisecond)

	// Release Ctrl
	err = ti.sendKeyEvent(VK_CONTROL, true)
	if err != nil {
		return fmt.Errorf("failed to release Ctrl: %w", err)
	}

	return nil
}

// Sleep provides a Windows-compatible sleep function
func (ti *TextInjector) Sleep(milliseconds uint32) {
	procSleep.Call(uintptr(milliseconds))
}
