// File: internal/ui/update_dialogs.go
package ui

import (
	"context"
	"fmt"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/internal/update"
)

// GitHub repository info
const (
	githubOwner = "Sarwarhridoy4"
	githubRepo  = "FyClip---Advanced-Clipboard-Manager"
)

// ShowUpdateDialog checks for updates and shows a dialog with the result
func ShowUpdateDialog(window fyne.Window, app fyne.App, currentVersion string) {
	updateWindow := app.NewWindow("Check for Updates")
	updateWindow.Resize(fyne.NewSize(450, 300))
	updateWindow.SetFixedSize(true)

	// Progress indicator
	progress := widget.NewProgressBar()
	progress.Hide()

	// Status label
	statusLabel := widget.NewLabel("Checking for updates...")
	statusLabel.Alignment = fyne.TextAlignCenter

	// Current version label
	versionLabel := widget.NewLabel(fmt.Sprintf("Current version: %s", currentVersion))
	versionLabel.Alignment = fyne.TextAlignCenter
	versionLabel.TextStyle.Italic = true

	// Result area - will be populated after check
	resultLabel := widget.NewLabel("")
	resultLabel.Alignment = fyne.TextAlignCenter
	resultLabel.Wrapping = fyne.TextWrapWord

	// Download button (initially hidden)
	downloadBtn := widget.NewButton("Download Update", func() {})
	downloadBtn.Hide()

	// Install button (initially hidden)
	installBtn := widget.NewButton("Install Update", func() {})
	installBtn.Hide()

	// Close button
	closeBtn := widget.NewButton("Close", func() {
		updateWindow.Close()
	})
	closeBtn.Importance = widget.HighImportance

	// Run update check in background
	go func() {
		checker := update.NewChecker(githubOwner, githubRepo, currentVersion)
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		updateInfo, err := checker.CheckForUpdate(ctx)

		fyne.Do(func() {
			progress.Hide()
			statusLabel.Hide()

			if err != nil {
				resultLabel.SetText(fmt.Sprintf("Error checking for updates:\n%v\n\nMake sure you have an internet connection.", err))
				updateWindow.Resize(fyne.NewSize(450, 250))
				return
			}

			comparison := update.CompareVersions(updateInfo.LatestVersion, currentVersion)
			if comparison > 0 {
				// Update available
				resultLabel.SetText(fmt.Sprintf("🎉 Update Available!\n\nLatest version: %s\nYour version: %s\n\n%s", 
					updateInfo.LatestVersion, currentVersion, updateInfo.ReleaseNotes))
				
				// Show download button
				downloadBtn.Show()
				downloadBtn.OnTapped = func() {
					ShowDownloadProgressDialog(updateWindow, app, checker, updateInfo, currentVersion)
				}

				updateWindow.Resize(fyne.NewSize(450, 400))
			} else {
				// No update available
				resultLabel.SetText(fmt.Sprintf("✅ You are using the latest version!\n\nVersion: %s", currentVersion))
			}
		})
	}()

	// Content
	content := container.NewVBox(
		container.NewHBox(layout.NewSpacer(), widget.NewIcon(theme.ViewRefreshIcon()), layout.NewSpacer()),
		container.NewPadded(versionLabel),
		container.NewPadded(statusLabel),
		progress,
		container.NewPadded(resultLabel),
		container.NewHBox(layout.NewSpacer(), downloadBtn, installBtn, layout.NewSpacer()),
		widget.NewSeparator(),
		closeBtn,
	)

	scrollContent := container.NewScroll(content)
	scrollContent.SetMinSize(fyne.NewSize(450, 300))

	updateWindow.SetContent(scrollContent)
	updateWindow.CenterOnScreen()
	updateWindow.Show()
}

// ShowDownloadProgressDialog shows progress of downloading update
func ShowDownloadProgressDialog(window fyne.Window, app fyne.App, checker *update.Checker, updateInfo *update.UpdateInfo, currentVersion string) {
	progressWindow := app.NewWindow("Downloading Update")
	progressWindow.Resize(fyne.NewSize(450, 200))
	progressWindow.SetFixedSize(true)

	// Progress bar
	progress := widget.NewProgressBar()
	progress.Min = 0
	progress.Max = 1

	// Status label
	statusLabel := widget.NewLabel("Preparing download...")
	statusLabel.Alignment = fyne.TextAlignCenter

	// Size label
	sizeLabel := widget.NewLabel("")
	sizeLabel.Alignment = fyne.TextAlignCenter

	// Download button (hidden initially)
	downloadBtn := widget.NewButton("Download", func() {})
	downloadBtn.Hide()

	// Install button (hidden initially)
	installBtn := widget.NewButton("Install Now", func() {})
	installBtn.Hide()

	// Cancel button
	cancelBtn := widget.NewButton("Cancel", func() {
		progressWindow.Close()
	})

	// Content
	content := container.NewVBox(
		container.NewHBox(layout.NewSpacer(), widget.NewIcon(theme.DownloadIcon()), layout.NewSpacer()),
		widget.NewLabel(fmt.Sprintf("Downloading %s", updateInfo.AssetName)),
		container.NewPadded(progress),
		container.NewPadded(statusLabel),
		sizeLabel,
		container.NewHBox(layout.NewSpacer(), downloadBtn, installBtn, layout.NewSpacer()),
		container.NewHBox(layout.NewSpacer(), cancelBtn, layout.NewSpacer()),
	)

	progressWindow.SetContent(content)
	progressWindow.CenterOnScreen()
	progressWindow.Show()

	// Start download
	go func() {
		downloader := update.NewDownloader(checker, updateInfo)
		
		err := downloader.Download(context.Background(), func(downloaded, total int64) {
			fyne.Do(func() {
				if total > 0 {
					progress.SetValue(float64(downloaded) / float64(total))
					statusLabel.SetText(fmt.Sprintf("%.1f%% downloaded", float64(downloaded)/float64(total)*100))
					sizeLabel.SetText(fmt.Sprintf("%s / %s", formatBytes(downloaded), formatBytes(total)))
				}
			})
		})

		fyne.Do(func() {
			if err != nil {
				statusLabel.SetText(fmt.Sprintf("Download failed: %v", err))
				log.Printf("Download error: %v", err)
				return
			}

			// Download complete - show install options
			statusLabel.SetText("Download complete!")
			sizeLabel.SetText(fmt.Sprintf("Downloaded to: %s", downloader.GetDownloadPath()))
			progress.SetValue(1)
			
			downloadBtn.Show()
			downloadBtn.OnTapped = func() {
				// Open file location
				log.Printf("Update downloaded to: %s", downloader.GetDownloadPath())
				ShowNotification(app, fmt.Sprintf("Update downloaded to: %s", downloader.GetDownloadPath()))
			}

			installBtn.Show()
			installBtn.OnTapped = func() {
				progressWindow.Close()
				ShowInstallConfirmation(progressWindow, app, downloader.GetDownloadPath())
			}
		})
	}()
}

// ShowInstallConfirmation shows confirmation before installing
func ShowInstallConfirmation(window fyne.Window, app fyne.App, downloadPath string) {
	dialog.ShowConfirm("Install Update", 
		fmt.Sprintf("The update has been downloaded to:\n%s\n\nDo you want to install it now? You may need to restart the application after installation.", downloadPath),
		func(confirmed bool) {
			if !confirmed {
				return
			}

			installer := update.NewInstaller(downloadPath, "FyClip")
			if err := installer.Install(); err != nil {
				ShowNotification(app, fmt.Sprintf("Installation failed: %v", err))
				log.Printf("Installation error: %v", err)
				return
			}

			ShowNotification(app, "Update installed! Please restart the application.")
		}, window)
}

// formatBytes formats bytes into human readable string
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}