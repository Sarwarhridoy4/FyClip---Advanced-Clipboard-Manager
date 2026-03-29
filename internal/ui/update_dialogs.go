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
	log.Printf("Opening download progress window for %s", updateInfo.AssetName)
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
		log.Printf("Starting download goroutine...")
		downloader := update.NewDownloader(checker, updateInfo)
		log.Printf("Downloader created, starting download...")
		
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

			log.Printf("Download completed successfully")
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
				ShowInstallConfirmation(window, app, downloader.GetDownloadPath())
			}
		})
	}()
}

// ShowInstallConfirmation shows confirmation before installing
func ShowInstallConfirmation(window fyne.Window, app fyne.App, downloadPath string) {
	var confirmDialog *dialog.ConfirmDialog
	confirmDialog = dialog.NewConfirm("Install Update", 
		fmt.Sprintf("The update has been downloaded to:\n%s\n\nDo you want to install it now? You may need to restart the application after installation.", downloadPath),
		func(confirmed bool) {
			if !confirmed {
				return
			}

			// Close the confirmation dialog before showing progress
			confirmDialog.Hide()
			ShowInstallProgressDialog(window, app, downloadPath)
		}, window)
	confirmDialog.Show()
}

// ShowInstallProgressDialog shows installation progress and logs
func ShowInstallProgressDialog(window fyne.Window, app fyne.App, downloadPath string) {
	installWindow := app.NewWindow("Installing Update")
	installWindow.Resize(fyne.NewSize(500, 400))
	installWindow.SetFixedSize(true)

	// Status label
	statusLabel := widget.NewLabel("Installing update...")
	statusLabel.Alignment = fyne.TextAlignCenter

	// Log output area
	logOutput := widget.NewMultiLineEntry()
	logOutput.SetText("Starting installation...\n")
	logOutput.Wrapping = fyne.TextWrapWord
	logOutput.Disable() // Make it read-only

	// Scroll container for logs
	logScroll := container.NewScroll(logOutput)
	logScroll.SetMinSize(fyne.NewSize(480, 250))

	// Close button (initially hidden)
	closeBtn := widget.NewButton("Close", func() {
		installWindow.Close()
	})
	closeBtn.Hide()

	// Content
	content := container.NewVBox(
		container.NewHBox(layout.NewSpacer(), widget.NewIcon(theme.DownloadIcon()), layout.NewSpacer()),
		container.NewPadded(statusLabel),
		widget.NewLabel("Installation Log:"),
		logScroll,
		container.NewHBox(layout.NewSpacer(), closeBtn, layout.NewSpacer()),
	)

	installWindow.SetContent(content)
	installWindow.CenterOnScreen()
	installWindow.Show()

	// Run installation in background
	go func() {
		installer := update.NewInstaller(downloadPath, "FyClip")
		
		// Update log periodically
		logChan := make(chan string, 100)
		go func() {
			for msg := range logChan {
				fyne.Do(func() {
					currentText := logOutput.Text
					logOutput.SetText(currentText + msg + "\n")
					logScroll.ScrollToBottom()
				})
			}
		}()

		logChan <- "Running installer..."
		logChan <- fmt.Sprintf("Download path: %s", downloadPath)
		
		err := installer.Install()
		
		// Get installation output
		output := installer.GetOutput()
		if output != "" {
			logChan <- "\n--- Installation Output ---"
			logChan <- output
		}
		
		close(logChan)

		fyne.Do(func() {
			if err != nil {
				statusLabel.SetText("Installation failed!")
				logOutput.SetText(logOutput.Text + fmt.Sprintf("\nError: %v", err))
				log.Printf("Installation error: %v", err)
			} else {
				statusLabel.SetText("Installation completed successfully!")
				logOutput.SetText(logOutput.Text + "\nUpdate installed! Please restart the application.")
			}
			closeBtn.Show()
		})
	}()
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