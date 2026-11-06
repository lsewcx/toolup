package main

import (
	"log/slog"
	"os"

	"github.com/lsewcx/toolup/internal/logging"
)

func main() {
	logging.Setup(os.Getenv("TOOLUP_DEBUG") != "")

	slog.Info("starting", "version", "0.1.0")
	slog.Debug("resolved backend", "tool", "docker", "backend", "dnf")
}
