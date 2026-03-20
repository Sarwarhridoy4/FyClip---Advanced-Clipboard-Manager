// File: internal/ui/toolbar.go
package ui

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/internal/clipboard"
)

// Toolbar provides action buttons.
type Toolbar struct {
	window  fyne.Window
	app     fyne.App
	manager *clipboard.Manager
	list    *HistoryList

	container    *fyne.Container
	favoritesBtn *widget.Button
	pauseBtn     *widget.Button
}

// NewToolbar creates a new toolbar.
func NewToolbar(window fyne.Window, app fyne.App, manager *clipboard.Manager, list *HistoryList) *Toolbar {
	return &Toolbar{
		window:  window,
		app:     app,
		manager: manager,
		list:    list,
	}
}

// Build creates the toolbar widget.
func (t *Toolbar) Build() fyne.CanvasObject {
	copyBtn := widget.NewButtonWithIcon("Copy", theme.ContentCopyIcon(), t.onCopy)
	pinBtn := widget.NewButtonWithIcon("Pin/Unpin", theme.ViewRefreshIcon(), t.onPin)
	t.favoritesBtn = widget.NewButtonWithIcon("", theme.ConfirmIcon(), t.onFavorites)
	t.pauseBtn = widget.NewButtonWithIcon("", theme.MediaPauseIcon(), t.onPause)
	deleteBtn := widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), t.onDelete)
	clearBtn := widget.NewButtonWithIcon("Clear", theme.ContentClearIcon(), t.onClear)
	snippetsBtn := widget.NewButtonWithIcon("Snippets", theme.FolderOpenIcon(), t.onSnippets)
	_ = snippetsBtn // Silence unused warning
	settingsBtn := widget.NewButtonWithIcon("Settings", theme.SettingsIcon(), t.onSettings)
	exportBtn := widget.NewButtonWithIcon("Export", theme.DocumentSaveIcon(), t.onExport)
	refreshBtn := widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), t.onRefresh)
	backupBtn := widget.NewButtonWithIcon("Backup", theme.DocumentSaveIcon(), t.onBackup)
	restoreBtn := widget.NewButtonWithIcon("Restore", theme.DocumentIcon(), t.onRestore)

	// First row: primary actions
	row1 := container.NewHBox(
		copyBtn,
		pinBtn,
		t.favoritesBtn,
		t.pauseBtn,
		deleteBtn,
		clearBtn,
	)

	// Second row: secondary actions
	row2 := container.NewHBox(
		snippetsBtn,
		settingsBtn,
		exportBtn,
		refreshBtn,
		backupBtn,
		restoreBtn,
	)

	t.container = container.NewVBox(
		row1,
		row2,
	)
	t.refreshToggleLabels()
	return t.container
}

func (t *Toolbar) refreshToggleLabels() {
	if t.favoritesBtn != nil {
		if t.manager.IsPinnedOnly() {
			t.favoritesBtn.SetText("All Items")
		} else {
			t.favoritesBtn.SetText("Favorites")
		}
	}
	if t.pauseBtn != nil {
		if t.manager.IsMonitoringPaused() {
			t.pauseBtn.SetText("Resume")
			t.pauseBtn.SetIcon(theme.MediaPlayIcon())
		} else {
			t.pauseBtn.SetText("Pause 5m")
			t.pauseBtn.SetIcon(theme.MediaPauseIcon())
		}
	}
}

// Refresh updates dynamic toolbar labels.
func (t *Toolbar) Refresh() {
	t.refreshToggleLabels()
}

// onCopy handles copy button.
func (t *Toolbar) onCopy() {
	index := t.manager.GetSelectedIndex()
	if index < 0 {
		ShowNotification(t.app, "No item selected!")
		return
	}

	if err := t.manager.CopyToClipboard(index); err != nil {
		ShowNotification(t.app, fmt.Sprintf("Copy failed: %v", err))
		return
	}

	ShowNotification(t.app, "Copied to clipboard!")
}

// onPin handles pin button.
func (t *Toolbar) onPin() {
	index := t.manager.GetSelectedIndex()
	if index < 0 {
		ShowNotification(t.app, "No item selected!")
		return
	}

	item, ok := t.manager.GetItem(index)
	if !ok {
		ShowNotification(t.app, "Failed to get item!")
		return
	}

	currentlyPinned := item.Pinned

	// Use FindIndexByID to get the current index to avoid stale index issues
	currentIndex := t.manager.FindIndexByID(item.ID)
	if currentIndex < 0 {
		ShowNotification(t.app, "Item not found!")
		return
	}

	if t.manager.TogglePin(currentIndex) {
		t.manager.SaveHistory()
		if currentlyPinned {
			ShowNotification(t.app, "Item unpinned!")
		} else {
			ShowNotification(t.app, "Item pinned!")
		}
		if t.list != nil {
			t.list.Refresh()
		}
	} else {
		ShowNotification(t.app, "Failed to toggle pin!")
	}
}

func (t *Toolbar) onFavorites() {
	enabled := t.manager.TogglePinnedOnly()
	if t.list != nil {
		t.list.UnselectAll()
		t.list.Refresh()
	}
	t.refreshToggleLabels()
	if enabled {
		ShowNotification(t.app, "Showing favorites only")
	} else {
		ShowNotification(t.app, "Showing all items")
	}
}

func (t *Toolbar) onPause() {
	if t.manager.IsMonitoringPaused() {
		t.manager.ResumeMonitoring()
		ShowNotification(t.app, "Clipboard monitoring resumed")
	} else {
		t.manager.PauseMonitoringFor(5 * time.Minute)
		ShowNotification(t.app, "Clipboard monitoring paused for 5 minutes")
	}
	t.refreshToggleLabels()
}

// onDelete handles delete button.
func (t *Toolbar) onDelete() {
	index := t.manager.GetSelectedIndex()
	if index < 0 {
		ShowNotification(t.app, "No item selected!")
		return
	}

	if err := t.manager.Delete(index); err != nil {
		ShowNotification(t.app, err.Error())
		return
	}

	t.manager.SaveHistory()
	if t.list != nil {
		t.list.UnselectAll()
		t.list.Refresh()
	}
	ShowNotification(t.app, "Item deleted!")
}

// onClear handles clear button.
func (t *Toolbar) onClear() {
	dialog.ShowConfirm(
		"Clear History",
		"Are you sure you want to clear all unpinned items?",
		func(confirmed bool) {
			if !confirmed {
				return
			}

			t.manager.ClearUnpinned()
			t.manager.SaveHistory()
			if t.list != nil {
				t.list.UnselectAll()
				t.list.Refresh()
			}
			ShowNotification(t.app, "History cleared!")
		},
		t.window,
	)
}

// onSnippets opens snippets management dialog
func (t *Toolbar) onSnippets() {
	snippets := t.manager.GetSnippets()
	
	// Create a list of snippets
	list := widget.NewList(
		func() int { return len(snippets) },
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id < len(snippets) {
				s := snippets[id]
				title := s.Title
				if s.Abbreviation != "" {
					title = title + " (" + s.Abbreviation + ")"
				}
				item.(*widget.Label).SetText(title)
			}
		},
	)

	// Show info
	infoLabel := widget.NewLabel("Snippets - Click to expand and copy")
	infoLabel.Alignment = fyne.TextAlignCenter

	content := container.NewVBox(
		infoLabel,
		widget.NewSeparator(),
		container.NewMax(list),
	)

	dialog.ShowCustom("Snippets", "Close", content, t.window)
}

// onSettings opens settings dialog
func (t *Toolbar) onSettings() {
	options := []string{"100", "500", "1000"}
	radio := widget.NewRadioGroup(options, nil)
	current := strconv.Itoa(t.manager.GetMaxHistory())
	radio.SetSelected(current)
	if radio.Selected == "" {
		radio.SetSelected("1000")
	}

	content := container.NewVBox(
		widget.NewLabel("Maximum unpinned history items"),
		radio,
	)

	dialog.ShowCustomConfirm("Settings", "Save", "Cancel", content, func(confirmed bool) {
		if !confirmed {
			return
		}
		limit, err := strconv.Atoi(radio.Selected)
		if err != nil || !t.manager.SetMaxHistory(limit) {
			ShowNotification(t.app, "Invalid history limit")
			return
		}
		ShowNotification(t.app, fmt.Sprintf("Max history set to %d", limit))
	}, t.window)
}

func (t *Toolbar) onExport() {
	item, ok := t.manager.GetSelected()
	if !ok {
		ShowNotification(t.app, "No item selected!")
		return
	}

	switch item.Type {
	case clipboard.TypeText:
		t.exportText(item)
	case clipboard.TypeImage:
		t.exportImage(item)
	default:
		ShowNotification(t.app, "Unsupported item type")
	}
}

func (t *Toolbar) exportText(item clipboard.Item) {
	suggested := fmt.Sprintf("clipboard_%s.txt", time.Now().Format("20060102_150405"))
	fileSaveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil || writer == nil {
			return
		}
		defer writer.Close()

		if _, err := writer.Write([]byte(item.Content)); err != nil {
			ShowNotification(t.app, fmt.Sprintf("Export failed: %v", err))
			return
		}
		ShowNotification(t.app, "Text exported")
	}, t.window)
	fileSaveDialog.SetFileName(suggested)
	fileSaveDialog.Show()
}

func (t *Toolbar) exportImage(item clipboard.Item) {
	ShowImageFormatDialog(t.window, func(selectedFormat string, err error) {
		if err != nil {
			return
		}

		defaultExtension := ".png"
		if selectedFormat == "jpeg" {
			defaultExtension = ".jpeg"
		}
		suggestedFilename := fmt.Sprintf("image_%s%s", time.Now().Format("20060102_150405"), defaultExtension)

		fileSaveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil || writer == nil {
				return
			}
			defer writer.Close()

			filename := writer.URI().Path()
			if !strings.HasSuffix(strings.ToLower(filename), ".png") &&
				!strings.HasSuffix(strings.ToLower(filename), ".jpg") &&
				!strings.HasSuffix(strings.ToLower(filename), ".jpeg") {
				filename += defaultExtension
			}

			formatToSave := selectedFormat
			ext := strings.ToLower(filepath.Ext(filename))
			switch ext {
			case ".jpg", ".jpeg":
				formatToSave = "jpeg"
			case ".png":
				formatToSave = "png"
			}

			if err := SaveImage(item, filename, formatToSave); err != nil {
				ShowNotification(t.app, fmt.Sprintf("Export failed: %v", err))
				return
			}
			ShowNotification(t.app, fmt.Sprintf("Saved as %s", filepath.Base(filename)))
		}, t.window)
		fileSaveDialog.SetFileName(suggestedFilename)
		fileSaveDialog.Show()
	})
}

// onRefresh handles refresh button.
func (t *Toolbar) onRefresh() {
	if t.list != nil {
		t.list.Refresh()
	}
	ShowNotification(t.app, "Clipboard history refreshed!")
}

// onBackup handles backup button - exports clipboard history to a backup file
func (t *Toolbar) onBackup() {
	backupMgr := t.manager.GetBackupManager()
	if backupMgr == nil {
		ShowNotification(t.app, "Backup manager not available")
		return
	}

	// Create password entry
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.PlaceHolder = "Optional password for encryption"

	// Create confirm password entry
	confirmPasswordEntry := widget.NewPasswordEntry()
	confirmPasswordEntry.PlaceHolder = "Confirm password"

	// Create checkbox for encryption
	encryptCheck := widget.NewCheck("Encrypt backup with password", nil)

	// Create form content
	formContent := container.NewVBox(
		widget.NewLabel("Export clipboard history to a backup file."),
		widget.NewSeparator(),
		encryptCheck,
		widget.NewLabel("Password:"),
		passwordEntry,
		widget.NewLabel("Confirm Password:"),
		confirmPasswordEntry,
	)

	// Show dialog
	dialog.ShowCustomConfirm("Backup Clipboard History", "Export", "Cancel", formContent, func(confirmed bool) {
		if !confirmed {
			return
		}

		password := ""
		if encryptCheck.Checked {
			password = passwordEntry.Text
			if password != confirmPasswordEntry.Text {
				ShowNotification(t.app, "Passwords do not match!")
				return
			}
			if password == "" {
				ShowNotification(t.app, "Password cannot be empty when encryption is enabled!")
				return
			}
		}

		// Show file save dialog
		suggested := fmt.Sprintf("fyclip_backup_%s.json", time.Now().Format("20060102_150405"))
		if encryptCheck.Checked {
			suggested = fmt.Sprintf("fyclip_backup_%s.enc", time.Now().Format("20060102_150405"))
		}

		fileSaveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil || writer == nil {
				return
			}
			defer writer.Close()

			path := writer.URI().Path()
			if err := backupMgr.ExportBackup(path, password); err != nil {
				ShowNotification(t.app, fmt.Sprintf("Backup failed: %v", err))
				return
			}

			ShowNotification(t.app, "Backup created successfully!")
		}, t.window)
		fileSaveDialog.SetFileName(suggested)
		fileSaveDialog.Show()
	}, t.window)
}

// onRestore handles restore button - imports clipboard history from a backup file
func (t *Toolbar) onRestore() {
	backupMgr := t.manager.GetBackupManager()
	if backupMgr == nil {
		ShowNotification(t.app, "Backup manager not available")
		return
	}

	// Show file open dialog
	fileOpenDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		defer reader.Close()

		path := reader.URI().Path()

		// Try to get backup info first
		info, err := backupMgr.GetBackupInfo(path)
		if err != nil {
			ShowNotification(t.app, fmt.Sprintf("Failed to read backup: %v", err))
			return
		}

		// Check if encrypted
		isEncrypted := info.Version == "encrypted"

		// Create password entry
		passwordEntry := widget.NewPasswordEntry()
		passwordEntry.PlaceHolder = "Enter password if backup is encrypted"

		// Create merge checkbox
		mergeCheck := widget.NewCheck("Merge with existing history (uncheck to replace)", nil)
		mergeCheck.SetChecked(true)

		// Create form content
		formContent := container.NewVBox(
			widget.NewLabel(fmt.Sprintf("Restore from backup file: %s", filepath.Base(path))),
			widget.NewLabel(fmt.Sprintf("Backup date: %s", info.Timestamp.Format("2006-01-02 15:04:05"))),
			widget.NewLabel(fmt.Sprintf("Items in backup: %d", len(info.Items))),
			widget.NewSeparator(),
		)

		if isEncrypted {
			formContent.Add(widget.NewLabel("This backup is encrypted. Enter password:"))
			formContent.Add(passwordEntry)
		}

		formContent.Add(mergeCheck)

		// Show confirmation dialog
		dialog.ShowCustomConfirm("Restore Clipboard History", "Restore", "Cancel", formContent, func(confirmed bool) {
			if !confirmed {
				return
			}

			password := ""
			if isEncrypted {
				password = passwordEntry.Text
				if password == "" {
					ShowNotification(t.app, "Password required for encrypted backup!")
					return
				}
			}

			merge := mergeCheck.Checked

			// Perform restore
			if err := backupMgr.ImportBackup(path, password, merge); err != nil {
				ShowNotification(t.app, fmt.Sprintf("Restore failed: %v", err))
				return
			}

			// Reload history
			if err := t.manager.ReloadHistory(); err != nil {
				ShowNotification(t.app, fmt.Sprintf("Failed to reload history: %v", err))
				return
			}

			if t.list != nil {
				t.list.UnselectAll()
				t.list.Refresh()
			}

			if merge {
				ShowNotification(t.app, "Backup merged successfully!")
			} else {
				ShowNotification(t.app, "Backup restored successfully!")
			}
		}, t.window)
	}, t.window)

	fileOpenDialog.SetFilter(storage.NewExtensionFileFilter([]string{".json", ".enc"}))
	fileOpenDialog.Show()
}
