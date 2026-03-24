// File: internal/ui/toolbar.go
package ui

import (
	"fmt"
	"image/color"
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

// forcedVariantTheme is a custom theme that wraps the default theme
// but forces a specific variant (light or dark) instead of following system preference.
type forcedVariantTheme struct {
	fyne.Theme
	variant fyne.ThemeVariant
}

// Color overrides the default Color method to return colors for the forced variant.
func (f forcedVariantTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return f.Theme.Color(name, f.variant)
}

// Variant returns the forced theme variant.
func (f forcedVariantTheme) Variant() fyne.ThemeVariant {
	return f.variant
}

// Toolbar provides action buttons.
type Toolbar struct {
	window  fyne.Window
	app     fyne.App
	manager *clipboard.Manager
	list    *HistoryList

	container      *fyne.Container
	favoritesBtn   *widget.Button
	pauseBtn       *widget.Button
	selectModeBtn   *widget.Button
	bulkDeleteBtn  *widget.Button
	bulkPinBtn     *widget.Button
	bulkUnpinBtn   *widget.Button
	bulkSelectAll  *widget.Button
	bulkClearSel   *widget.Button
	themeBtn       *widget.Button
	themeMenu      *fyne.Menu
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
	pinBtn := widget.NewButtonWithIcon("Pin", theme.ConfirmIcon(), t.onPin)
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

	// Theme selector with icons using a button that opens a popup menu
	t.themeBtn = widget.NewButtonWithIcon("Theme", theme.ColorPaletteIcon(), t.onThemeButtonClicked)
	// Create the theme popup menu
	t.createThemeMenu()

	// Selection mode buttons
	t.selectModeBtn = widget.NewButtonWithIcon("Select", theme.CheckButtonIcon(), t.onToggleSelectMode)
	t.bulkSelectAll = widget.NewButtonWithIcon("All", theme.MenuIcon(), t.onSelectAll)
	t.bulkClearSel = widget.NewButtonWithIcon("None", theme.CancelIcon(), t.onClearSelection)
	t.bulkPinBtn = widget.NewButtonWithIcon("Pin", theme.ConfirmIcon(), t.onBulkPin)
	t.bulkUnpinBtn = widget.NewButtonWithIcon("Unpin", theme.RadioButtonIcon(), t.onBulkUnpin)
	t.bulkDeleteBtn = widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), t.onBulkDelete)

	// Hide bulk action buttons initially
	t.bulkSelectAll.Hide()
	t.bulkClearSel.Hide()
	t.bulkPinBtn.Hide()
	t.bulkUnpinBtn.Hide()
	t.bulkDeleteBtn.Hide()

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
		t.themeBtn,
		settingsBtn,
		exportBtn,
		refreshBtn,
		backupBtn,
		restoreBtn,
	)

	// Third row: selection mode and bulk actions
	row3 := container.NewHBox(
		t.selectModeBtn,
		t.bulkSelectAll,
		t.bulkClearSel,
		widget.NewSeparator(),
		t.bulkPinBtn,
		t.bulkUnpinBtn,
		t.bulkDeleteBtn,
	)

	t.container = container.NewVBox(
		row1,
		row2,
		row3,
	)
	t.refreshToggleLabels()
	return t.container
}

func (t *Toolbar) refreshToggleLabels() {
	if t.favoritesBtn != nil {
		if t.manager.IsPinnedOnly() {
			t.favoritesBtn.SetText("Show All")
			t.favoritesBtn.SetIcon(theme.MenuIcon())
		} else {
			t.favoritesBtn.SetText("Pinned Only")
			t.favoritesBtn.SetIcon(theme.RadioButtonIcon())
		}
	}
	if t.pauseBtn != nil {
		if t.manager.IsMonitoringPaused() {
			t.pauseBtn.SetText("Resume")
			t.pauseBtn.SetIcon(theme.MediaPlayIcon())
		} else {
			t.pauseBtn.SetText("Pause")
			t.pauseBtn.SetIcon(theme.MediaPauseIcon())
		}
	}
}

// Refresh updates dynamic toolbar labels.
func (t *Toolbar) Refresh() {
	t.refreshToggleLabels()
	t.updateThemeButton()
}

// updateThemeButton updates the theme button text and icon based on current theme
func (t *Toolbar) updateThemeButton() {
	if t.app == nil {
		return
	}

	settings := t.app.Settings()
	if settings == nil {
		return
	}

	currentThemeObj := settings.Theme()

	// Determine current theme variant by checking colors
	// Default to system
	selected := "System"

	if currentThemeObj != nil {
		// Check if it's our custom forced variant theme
		if fvt, ok := currentThemeObj.(interface{ variant() fyne.ThemeVariant }); ok {
			v := fvt.variant()
			switch v {
			case theme.VariantLight:
				selected = "Light"
			case theme.VariantDark:
				selected = "Dark"
			}
		} else {
			// For other themes, default to System (follows system preference)
			selected = "System"
		}
	}

	if t.themeBtn != nil {
		t.themeBtn.SetText(selected)
		t.updateThemeButtonIcon(selected)
	}
}

// createThemeMenu creates the theme selection menu with icons
func (t *Toolbar) createThemeMenu() {
		if t.window == nil {
		return
	}

	// Create menu items - use available icons
	systemItem := fyne.NewMenuItem("System (Follow OS)", func() { t.setTheme("System") })
	systemItem.Icon = theme.ComputerIcon()

	darkItem := fyne.NewMenuItem("Dark", func() { t.setTheme("Dark") })
	darkItem.Icon = theme.VisibilityIcon()

	lightItem := fyne.NewMenuItem("Light", func() { t.setTheme("Light") })
	lightItem.Icon = theme.ColorPaletteIcon()

	t.themeMenu = fyne.NewMenu("", systemItem, darkItem, lightItem)
}

// onThemeButtonClicked shows the theme popup menu
func (t *Toolbar) onThemeButtonClicked() {
	if t.themeMenu != nil && t.window != nil {
		// Create a popup menu
		popup := widget.NewPopUpMenu(t.themeMenu, t.window.Canvas())

		// Position popup at center of window
		windowSize := t.window.Canvas().Size()
		menuSize := popup.MinSize()
		centerPos := fyne.NewPos(
			(windowSize.Width-menuSize.Width)/2,
			(windowSize.Height-menuSize.Height)/2,
		)
		popup.ShowAtPosition(centerPos)
	}
}

// setTheme sets the application theme
func (t *Toolbar) setTheme(themeName string) {
	if t.app == nil {
		return
	}

	settings := t.app.Settings()
	if settings == nil {
		return
	}

	switch themeName {
	case "Light":
		settings.SetTheme(forcedVariantTheme{
			Theme:   theme.DefaultTheme(),
			variant: theme.VariantLight,
		})
		ShowNotification(t.app, "Theme set to Light")
	case "Dark":
		settings.SetTheme(forcedVariantTheme{
			Theme:   theme.DefaultTheme(),
			variant: theme.VariantDark,
		})
		ShowNotification(t.app, "Theme set to Dark")
	case "System":
		// Reset to default theme to follow system preference
		settings.SetTheme(theme.DefaultTheme())
		ShowNotification(t.app, "Theme set to System (follows OS)")
	}

	// Update button appearance
	t.updateThemeButtonIcon(themeName)
	t.themeBtn.SetText(themeName)
}

// updateThemeButtonIcon updates the theme button icon based on current selection
func (t *Toolbar) updateThemeButtonIcon(themeName string) {
	if t.themeBtn == nil {
		return
	}

	var icon fyne.Resource
	switch themeName {
	case "Light":
		icon = theme.ColorPaletteIcon()
	case "Dark":
		icon = theme.VisibilityIcon()
	default:
		icon = theme.ComputerIcon()
	}

	t.themeBtn.SetIcon(icon)
}



// SetSelectionModeActive updates the toolbar UI based on selection mode state
func (t *Toolbar) SetSelectionModeActive(active bool) {
	if active {
		t.selectModeBtn.SetText("Cancel")
		t.bulkSelectAll.Show()
		t.bulkClearSel.Show()
		t.bulkPinBtn.Show()
		t.bulkUnpinBtn.Show()
		t.bulkDeleteBtn.Show()
	} else {
		t.selectModeBtn.SetText("Select")
		t.bulkSelectAll.Hide()
		t.bulkClearSel.Hide()
		t.bulkPinBtn.Hide()
		t.bulkUnpinBtn.Hide()
		t.bulkDeleteBtn.Hide()
	}
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
		ShowNotification(t.app, "Showing pinned items only")
	} else {
		ShowNotification(t.app, "Showing all items")
	}
}

func (t *Toolbar) onPause() {
	if t.manager.IsMonitoringPaused() {
		t.manager.PauseMonitoringFor(5 * time.Minute)
		ShowNotification(t.app, "Clipboard monitoring paused for 5 minutes")
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

// onToggleSelectMode toggles multi-select mode
func (t *Toolbar) onToggleSelectMode() {
	if t.list == nil {
		return
	}

	currentMode := t.list.IsSelectionMode()
	newMode := !currentMode
	t.list.SetSelectionMode(newMode)

	if newMode {
		ShowNotification(t.app, "Selection mode enabled - click items to select")
	} else {
		ShowNotification(t.app, "Selection mode disabled")
	}

	t.SetSelectionModeActive(newMode)
	t.list.Refresh()
}

// onSelectAll selects all visible items
func (t *Toolbar) onSelectAll() {
	if t.list == nil {
		return
	}
	t.list.SelectAll()
	count := t.list.GetSelectedCount()
	ShowNotification(t.app, fmt.Sprintf("%d items selected", count))
}

// onClearSelection clears the current selection
func (t *Toolbar) onClearSelection() {
	if t.list == nil {
		return
	}
	t.list.ClearSelection()
	ShowNotification(t.app, "Selection cleared")
}

// onBulkPin pins all selected items
func (t *Toolbar) onBulkPin() {
	if t.list == nil {
		return
	}

	ids := t.list.GetSelectedIDs()
	if len(ids) == 0 {
		ShowNotification(t.app, "No items selected!")
		return
	}

	count := t.manager.BulkPin(ids)
	t.manager.SaveHistory()
	t.list.ClearSelection()
	t.list.Refresh()
	ShowNotification(t.app, fmt.Sprintf("%d items pinned", count))
}

// onBulkUnpin unpins all selected items
func (t *Toolbar) onBulkUnpin() {
	if t.list == nil {
		return
	}

	ids := t.list.GetSelectedIDs()
	if len(ids) == 0 {
		ShowNotification(t.app, "No items selected!")
		return
	}

	count := t.manager.BulkUnpin(ids)
	t.manager.SaveHistory()
	t.list.ClearSelection()
	t.list.Refresh()
	ShowNotification(t.app, fmt.Sprintf("%d items unpinned", count))
}

// onBulkDelete deletes all selected items
func (t *Toolbar) onBulkDelete() {
	if t.list == nil {
		return
	}

	ids := t.list.GetSelectedIDs()
	if len(ids) == 0 {
		ShowNotification(t.app, "No items selected!")
		return
	}

	dialog.ShowConfirm(
		"Delete Selected Items",
		fmt.Sprintf("Are you sure you want to delete %d selected items? (Pinned items will be skipped)", len(ids)),
		func(confirmed bool) {
			if !confirmed {
				return
			}

			count := t.manager.BulkDelete(ids)
			t.manager.SaveHistory()
			t.list.ClearSelection()
			t.list.Refresh()
			ShowNotification(t.app, fmt.Sprintf("%d items deleted", count))
		},
		t.window,
	)
}

// onSnippets opens snippets management dialog
func (t *Toolbar) onSnippets() {
	// Use a local variable to track selection
	var selectedIndex int = -1
	snippets := t.manager.GetSnippets()
	
	// Create info label
	infoLabel := widget.NewLabel("Snippets - Manage your text templates")
	infoLabel.Alignment = fyne.TextAlignCenter
	
	// Variable reference label
	varLabel := widget.NewLabel("Variables: {{date}} {{time}} {{datetime}} {{clipboard}} {{year}} {{month}} {{day}}")
	varLabel.Alignment = fyne.TextAlignCenter
	varLabel.Wrapping = fyne.TextWrapWord
	varLabel.TextStyle.Italic = true
	
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
				if s.IsSystem {
					title = title + " [System]"
				}
				item.(*widget.Label).SetText(title)
			}
		},
	)
	
	// Preview area for selected snippet
	previewTitle := widget.NewLabel("")
	previewTitle.TextStyle.Bold = true
	previewContent := widget.NewLabel("")
	previewContent.Wrapping = fyne.TextWrapWord
	
	previewCard := widget.NewCard("", "", container.NewVBox(previewTitle, previewContent))
	
	// Update preview when selection changes
	list.OnSelected = func(id widget.ListItemID) {
		selectedIndex = int(id)
		if id >= 0 && id < len(snippets) {
			s := snippets[id]
			titleText := s.Title
			if s.Abbreviation != "" {
				titleText = s.Title + " (" + s.Abbreviation + ")"
			}
			previewTitle.SetText(titleText)
			// Expand template variables for preview
			clipboardContent := ""
			if fyne.CurrentApp() != nil && fyne.CurrentApp().Clipboard() != nil {
				clipboardContent = fyne.CurrentApp().Clipboard().Content()
			}
			expanded := t.manager.ExpandSnippetWithClipboard(s.Content, clipboardContent)
			previewContent.SetText(expanded)
		}
	}
	
	// Buttons
	useBtn := widget.NewButton("Use", func() {
		if selectedIndex < 0 || selectedIndex >= len(snippets) {
			ShowNotification(t.app, "No snippet selected!")
			return
		}
		s := snippets[selectedIndex]
		// Expand template variables
		clipboardContent := ""
		if fyne.CurrentApp() != nil && fyne.CurrentApp().Clipboard() != nil {
			clipboardContent = fyne.CurrentApp().Clipboard().Content()
		}
		expanded := t.manager.ExpandSnippetWithClipboard(s.Content, clipboardContent)
		fyne.CurrentApp().Clipboard().SetContent(expanded)
		ShowNotification(t.app, "Snippet copied to clipboard!")
	})
	
	addBtn := widget.NewButton("Add", func() {
		t.showAddSnippetDialog()
	})
	
	deleteBtn := widget.NewButton("Delete", func() {
		if selectedIndex < 0 || selectedIndex >= len(snippets) {
			ShowNotification(t.app, "No snippet selected!")
			return
		}
		s := snippets[selectedIndex]
		
		// Check if it's a system snippet
		if s.IsSystem {
			ShowNotification(t.app, "Cannot delete system snippet!")
			return
		}
		
		dialog.ShowConfirm("Delete Snippet", "Are you sure you want to delete \""+s.Title+"\"?", func(confirmed bool) {
			if confirmed {
				err := t.manager.DeleteSnippet(s.ID)
				if err != nil {
					ShowNotification(t.app, "Failed to delete snippet!")
					return
				}
				ShowNotification(t.app, "Snippet deleted!")
				// Refresh the dialog
				t.onSnippets()
			}
		}, t.window)
	})
	
	buttonBox := container.NewHBox(useBtn, addBtn, deleteBtn)
	
	content := container.NewVBox(
		infoLabel,
		varLabel,
		widget.NewSeparator(),
		container.NewMax(list),
		widget.NewSeparator(),
		previewCard,
		buttonBox,
	)

	dialog.ShowCustom("Snippets", "Close", content, t.window)
}

// showAddSnippetDialog shows a dialog to add a new snippet
func (t *Toolbar) showAddSnippetDialog() {
	// Title input
	titleEntry := widget.NewEntry()
	titleEntry.PlaceHolder = "Snippet title (e.g., Email Signature)"
	
	// Content input - using a large entry
	contentEntry := widget.NewEntry()
	contentEntry.PlaceHolder = "Snippet content - Use variables like {{date}}, {{clipboard}}, etc."
	
	// Abbreviation input
	abbrEntry := widget.NewEntry()
	abbrEntry.PlaceHolder = "Short abbreviation (optional, e.g., sig)"
	
	// Category input
	categoryEntry := widget.NewEntry()
	categoryEntry.PlaceHolder = "Category (optional, e.g., General)"
	
	form := container.NewVBox(
		widget.NewLabel("Title:"),
		titleEntry,
		widget.NewLabel("Abbreviation:"),
		abbrEntry,
		widget.NewLabel("Category:"),
		categoryEntry,
		widget.NewLabel("Content:"),
		contentEntry,
	)
	
	dialog.ShowCustomConfirm("Add Snippet", "Save", "Cancel", form, func(confirmed bool) {
		if !confirmed {
			return
		}
		
		if titleEntry.Text == "" {
			ShowNotification(t.app, "Title is required!")
			return
		}
		if contentEntry.Text == "" {
			ShowNotification(t.app, "Content is required!")
			return
		}
		
		snippet := clipboard.Snippet{
			Title:        titleEntry.Text,
			Content:      contentEntry.Text,
			Abbreviation: abbrEntry.Text,
			Category:     categoryEntry.Text,
		}
		
		err := t.manager.AddSnippet(snippet)
		if err != nil {
			ShowNotification(t.app, "Failed to save snippet!")
			return
		}
		
		ShowNotification(t.app, "Snippet saved!")
	}, t.window)
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
