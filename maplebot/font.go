package maplebot

import (
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	. "github.com/vicanso/go-charts/v2"
)

func init() {
	var fontDir string
	if runtime.GOOS == "windows" {
		fontDir = "C:\\Windows\\Fonts\\"
	} else if runtime.GOOS == "linux" {
		fontDir = "/usr/share/fonts/"
	} else if runtime.GOOS == "darwin" {
		fontDir = "/Library/Fonts/"
	} else {
		slog.Warn("Unsupported OS.")
		return
	}
	var fontFilePath string
	err := filepath.Walk(fontDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Base(path) == "simhei.ttf" {
			fontFilePath = path
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		slog.Warn("Failed to search font file.", "err", err)
		return
	}
	if len(fontFilePath) == 0 {
		slog.Warn("Font file not found.")
		return
	}
	// 加载字体文件
	fontBytes, err := os.ReadFile(fontFilePath)
	if err != nil {
		slog.Warn("Failed to read font file.", "err", err)
		return
	}
	err = InstallFont("simhei", fontBytes)
	if err != nil {
		slog.Warn("Failed to install font.", "err", err)
		return
	}
	font, _ := GetFont("simhei")
	SetDefaultFont(font)
}
