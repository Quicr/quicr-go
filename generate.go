// SPDX-FileCopyrightText: Copyright (c) 2025 Cisco Systems
// SPDX-License-Identifier: BSD-2-Clause

//go:build ignore

package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

const repo = "quicr/qgo"

func main() {
	if err := downloadShim(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func downloadShim() error {
	// Detect platform
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	artifact := fmt.Sprintf("qgo-libs-%s-%s.tar.gz", goos, goarch)
	fmt.Printf("Platform: %s-%s\n", goos, goarch)

	// Get latest release URL
	fmt.Println("Fetching latest release...")
	url, err := getLatestReleaseURL(artifact)
	if err != nil {
		return fmt.Errorf("could not find pre-built libraries for %s-%s: %w\nYou may need to build from source: make shim", goos, goarch, err)
	}

	fmt.Printf("Downloading: %s\n", url)

	// Create build directories
	dirs := []string{
		"build/lib",
		"build/libquicr/dependencies/picoquic",
		"build/libquicr/dependencies/spdlog",
		"build/_deps/picotls-build",
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// Download
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	// Extract
	fmt.Println("Extracting libraries...")
	if err := extractTarGz(resp.Body); err != nil {
		return err
	}

	fmt.Println("\nSuccessfully downloaded pre-built libraries!")
	fmt.Println("You can now build your project with: go build ./...")
	return nil
}

func getLatestReleaseURL(artifact string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var release struct {
		Assets []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	for _, asset := range release.Assets {
		if asset.Name == artifact {
			return asset.BrowserDownloadURL, nil
		}
	}

	return "", fmt.Errorf("artifact %s not found in release", artifact)
}

func extractTarGz(r io.Reader) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)

	// Map library files to their destinations
	libDest := map[string]string{
		"libquicr_shim.a":      "build/lib/",
		"libquicr.a":           "build/lib/",
		"libpicoquic-core.a":   "build/libquicr/dependencies/picoquic/",
		"libpicoquic-log.a":    "build/libquicr/dependencies/picoquic/",
		"libpicohttp-core.a":   "build/libquicr/dependencies/picoquic/",
		"libspdlog.a":          "build/libquicr/dependencies/spdlog/",
		"libpicotls-openssl.a": "build/_deps/picotls-build/",
		"libpicotls-core.a":    "build/_deps/picotls-build/",
		"libpicotls-minicrypto.a": "build/_deps/picotls-build/",
	}

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		name := filepath.Base(header.Name)
		dest, ok := libDest[name]
		if !ok {
			continue
		}

		outPath := filepath.Join(dest, name)
		outFile, err := os.Create(outPath)
		if err != nil {
			return err
		}

		if _, err := io.Copy(outFile, tr); err != nil {
			outFile.Close()
			return err
		}
		outFile.Close()

		fmt.Printf("  %s\n", outPath)
	}

	return nil
}
