package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/HP-jh/tars-update"
)

const (
	updateOwner = "HP-jh"
	updateRepo  = "TarsVoiceCorrection"
)

func runUpdateCheck(currentVersion string) {
	checker := update.NewChecker(updateOwner, updateRepo)
	release, err := checker.LatestRelease()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Update check failed:", err)
		fallbackToStable(currentVersion)
		return
	}

	latest, err := update.ParseSemver(release.TagName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to parse latest version:", err)
		fallbackToStable(currentVersion)
		return
	}

	current, _ := update.ParseSemver(currentVersion)
	if latest.IsNewerThan(current) {
		fmt.Printf("New version available: %s (current: %s)\n", release.TagName, currentVersion)
		if asset, ok := release.FindAssetForPlatform(); ok {
			fmt.Printf("Asset: %s (%d bytes)\n", asset.Name, asset.Size)
		}
		fmt.Printf("Release page: %s\n", release.HTMLURL)
		os.Exit(2) // conventional exit code for "update available"
	}
	fmt.Printf("You are up to date (%s)\n", currentVersion)
}

func runUpdateApply(currentVersion string) {
	checker := update.NewChecker(updateOwner, updateRepo)
	release, err := checker.LatestRelease()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Update check failed:", err)
		fallbackToStable(currentVersion)
		os.Exit(1)
	}

	latest, err := update.ParseSemver(release.TagName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to parse latest version:", err)
		fallbackToStable(currentVersion)
		os.Exit(1)
	}

	current, _ := update.ParseSemver(currentVersion)
	if !latest.IsNewerThan(current) {
		fmt.Printf("Already up to date (%s)\n", currentVersion)
		return
	}

	asset, ok := release.FindAssetForPlatform()
	if !ok {
		fmt.Fprintf(os.Stderr, "No asset found for platform %s/%s\n", runtime.GOOS, runtime.GOARCH)
		os.Exit(1)
	}

	// Download to temp dir
	tmpDir := filepath.Join(os.TempDir(), "tars-vc-update")
	downloader := update.NewDownloader()
	result, err := downloader.Download(asset, tmpDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Download failed:", err)
		os.Exit(1)
	}
	fmt.Printf("Downloaded %s (%d bytes)\n", result.Path, result.Size)

	// Try to verify SHA256 from SHA256SUMS.txt
	if sumsAsset, ok := release.FindAsset("SHA256SUMS"); ok {
		sumsResult, err := downloader.Download(sumsAsset, tmpDir)
		if err == nil {
			if sumsData, err := os.ReadFile(sumsResult.Path); err == nil {
				hashMap := update.ParseSHA256SUMS(string(sumsData))
				if expected, ok := hashMap[asset.Name]; ok {
					if err := update.VerifySHA256(result.Path, expected); err != nil {
						fmt.Fprintln(os.Stderr, "SHA256 verification failed:", err)
						os.Exit(1)
					}
					fmt.Println("SHA256 verification: OK")
				}
			}
		}
	}

	// Install
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot locate current executable:", err)
		os.Exit(1)
	}

	installer := update.NewInstaller()
	installResult, err := installer.Install(exe, result.Path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Install failed:", err)
		os.Exit(1)
	}
	fmt.Printf("Installed to %s (backup: %s)\n", installResult.NewPath, installResult.BackupPath)

	// Update stable_versions.json
	stable, _ := update.LoadStableVersions(update.DefaultStablePath())
	stable.Set("TarsVoiceCorrection", release.TagName)
	_ = stable.Save(update.DefaultStablePath())

	// Self-restart
	fmt.Println("Restarting...")
	if err := update.SelfRestart(); err != nil {
		fmt.Fprintln(os.Stderr, "Restart failed:", err)
		fmt.Println("Please restart manually.")
		os.Exit(1)
	}
	os.Exit(0)
}

func fallbackToStable(currentVersion string) {
	stable, err := update.LoadStableVersions(update.DefaultStablePath())
	if err != nil {
		return
	}
	v := stable.Get("TarsVoiceCorrection")
	if v == "" {
		return
	}
	fmt.Printf("Falling back to last known stable version: %s\n", v)
	if sv, err := update.ParseSemver(v); err == nil {
		cv, _ := update.ParseSemver(currentVersion)
		if sv.IsNewerThan(cv) {
			fmt.Println("Note: stable version is newer than current. Run 'update apply' to upgrade.")
		}
	}
}

func backgroundUpdateCheck(currentVersion string) {
	go func() {
		checker := update.NewChecker(updateOwner, updateRepo)
		checker.HTTPClient.Timeout = 10 * time.Second
		release, err := checker.LatestRelease()
		if err != nil {
			return // silently fail on background check
		}
		latest, err := update.ParseSemver(release.TagName)
		if err != nil {
			return
		}
		current, _ := update.ParseSemver(currentVersion)
		if latest.IsNewerThan(current) {
			fmt.Fprintf(os.Stderr, "\n[Update] New version %s is available (current: %s). Run '%s update apply' to upgrade.\n\n",
				release.TagName, currentVersion, filepath.Base(os.Args[0]))
		}
	}()
}
