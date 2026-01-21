// File: internal/clipboard/native.go
package clipboard

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"golang.design/x/clipboard"
)

// NativeClipboard handles platform-specific clipboard operations
type NativeClipboard struct {
	available bool
	useXclip  bool
	useWlclip bool
}

// NewNativeClipboard initializes clipboard support
func NewNativeClipboard() (*NativeClipboard, error) {
	nc := &NativeClipboard{
		available: true,
	}

	err := clipboard.Init()
	if err != nil {
		if runtime.GOOS == "linux" {
			nc.setupLinuxFallback()
		} else {
			return nil, fmt.Errorf("clipboard initialization failed: %w", err)
		}
	}

	return nc, nil
}

// setupLinuxFallback configures alternative clipboard methods for Linux
func (nc *NativeClipboard) setupLinuxFallback() {
	sessionType := os.Getenv("XDG_SESSION_TYPE")
	
	if sessionType == "wayland" || sessionType == "" {
		if _, err := exec.LookPath("wl-paste"); err == nil {
			nc.useWlclip = true
			log.Println("Using wl-clipboard for Wayland")
			return
		}
		log.Println("Warning: wl-clipboard not found for Wayland")
	}
	
	if _, err := exec.LookPath("xclip"); err == nil {
		nc.useXclip = true
		log.Println("Using xclip for X11")
		return
	}
	
	log.Println("Warning: No clipboard tool found. Install xclip or wl-clipboard")
	nc.available = false
}

// ReadText reads text from clipboard
func (nc *NativeClipboard) ReadText() []byte {
	if !nc.available {
		return nil
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic in ReadText: %v", r)
		}
	}()

	if nc.useWlclip {
		return nc.readTextWayland()
	} else if nc.useXclip {
		return nc.readTextX11()
	}

	return clipboard.Read(clipboard.FmtText)
}

// ReadImage reads image from clipboard
func (nc *NativeClipboard) ReadImage() ([]byte, string) {
	if !nc.available {
		return nil, ""
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic in ReadImage: %v", r)
		}
	}()

	if nc.useWlclip {
		return nc.readImageWayland()
	} else if nc.useXclip {
		return nc.readImageX11()
	}

	imgData := clipboard.Read(clipboard.FmtImage)
	if len(imgData) > 0 {
		return imgData, detectImageType(imgData)
	}

	return nil, ""
}

// WriteText writes text to clipboard
func (nc *NativeClipboard) WriteText(data []byte) error {
	if !nc.available {
		return fmt.Errorf("clipboard not available")
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic in WriteText: %v", r)
		}
	}()

	if nc.useWlclip {
		return nc.writeTextWayland(data)
	} else if nc.useXclip {
		return nc.writeTextX11(data)
	}

	clipboard.Write(clipboard.FmtText, data)
	return nil
}

// WriteImage writes image to clipboard
func (nc *NativeClipboard) WriteImage(base64Data string) error {
	if !nc.available {
		return fmt.Errorf("clipboard not available")
	}

	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic in WriteImage: %v", r)
		}
	}()

	if nc.useWlclip {
		return nc.writeImageWayland(data)
	} else if nc.useXclip {
		return nc.writeImageX11(data)
	}

	clipboard.Write(clipboard.FmtImage, data)
	return nil
}

// Linux Wayland implementations
func (nc *NativeClipboard) readTextWayland() []byte {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, "wl-paste", "-n")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	return out
}

func (nc *NativeClipboard) readImageWayland() ([]byte, string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, "wl-paste", "-t", "image/png", "-n")
	out, err := cmd.Output()
	if err == nil && len(out) > 0 {
		return out, "png"
	}
	
	ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel2()
	
	cmd2 := exec.CommandContext(ctx2, "wl-paste", "-t", "image/jpeg", "-n")
	out, err = cmd2.Output()
	if err == nil && len(out) > 0 {
		return out, "jpeg"
	}
	
	return nil, ""
}

func (nc *NativeClipboard) writeTextWayland(data []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, "wl-copy")
	cmd.Stdin = strings.NewReader(string(data))
	return cmd.Run()
}

func (nc *NativeClipboard) writeImageWayland(data []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, "wl-copy", "--type", "image/png")
	cmd.Stdin = strings.NewReader(string(data))
	return cmd.Run()
}

// Linux X11 implementations
func (nc *NativeClipboard) readTextX11() []byte {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, "xclip", "-o", "-selection", "clipboard")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	return out
}

func (nc *NativeClipboard) readImageX11() ([]byte, string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, "xclip", "-selection", "clipboard", "-t", "image/png", "-o")
	out, err := cmd.Output()
	if err == nil && len(out) > 0 {
		return out, "png"
	}
	
	ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel2()
	
	cmd2 := exec.CommandContext(ctx2, "xclip", "-selection", "clipboard", "-t", "image/jpeg", "-o")
	out, err = cmd2.Output()
	if err == nil && len(out) > 0 {
		return out, "jpeg"
	}
	
	return nil, ""
}

func (nc *NativeClipboard) writeTextX11(data []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, "xclip", "-i", "-selection", "clipboard")
	cmd.Stdin = strings.NewReader(string(data))
	return cmd.Run()
}

func (nc *NativeClipboard) writeImageX11(data []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, "xclip", "-selection", "clipboard", "-t", "image/png", "-i")
	cmd.Stdin = strings.NewReader(string(data))
	return cmd.Run()
}

// detectImageType detects the image type from raw data
func detectImageType(data []byte) string {
	if len(data) < 8 {
		return "png"
	}
	
	// PNG signature
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "png"
	}
	
	// JPEG signature
	if data[0] == 0xFF && data[1] == 0xD8 {
		return "jpeg"
	}
	
	return "png"
}

// IsAvailable returns whether clipboard is available
func (nc *NativeClipboard) IsAvailable() bool {
	return nc.available
}