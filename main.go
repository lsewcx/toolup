package main

import (
	"os"

	"github.com/lsewcx/toolup/cmd"
	"github.com/lsewcx/toolup/internal/logging"
)

func main() {
	logging.Setup(os.Getenv("TOOLUP_DEBUG") != "")
	cmd.Execute()
}
