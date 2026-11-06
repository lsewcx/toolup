package logging

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

// Setup 初始化全局日志器，在 main() 最开始调用一次即可。
func Setup(debug bool) {
	w := os.Stderr

	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}

	slog.SetDefault(slog.New(tint.NewTextHandler(w, &tint.Options{
		Level:      level,
		TimeFormat: "15:04:05",
		NoColor:    !useColor(w),
		AddSource:  debug,
	})))
}

// useColor 决定是否输出 ANSI 颜色。
func useColor(w *os.File) bool {
	if os.Getenv("NO_COLOR") != "" { // 遵守 no-color.org 约定
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	// IsCygwinTerminal 用来兼容 Windows 上的 Git Bash / mintty
	return isatty.IsTerminal(w.Fd()) || isatty.IsCygwinTerminal(w.Fd())
}
