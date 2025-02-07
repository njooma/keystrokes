package main

import (
	"keystrokes/models"
	"os"

	"github.com/viam-labs/screenshot-cam/subproc"
	"go.viam.com/rdk/components/generic"
	"go.viam.com/rdk/module"
	"go.viam.com/rdk/resource"
)

func main() {
	logger := module.NewLoggerFromArgs("screenshot-cam")
	var arg string
	if len(os.Args) >= 2 {
		arg = os.Args[1]
	}
	switch arg {
	case "parent":
		// parent is a test mode for spawning a child proc directly from session 0 CLI. see README.md for instructions.
		if err := subproc.SpawnSelf(" child"); err != nil {
			panic(err)
		}
	case "child":
		// child is the subprocess started in session 1 by a session 0 parent. it does the work.
		logger.Info("doing a keystrokes test instead of starting module")
	default:
		// ModularMain can take multiple APIModel arguments, if your module implements multiple models.
		module.ModularMain(resource.APIModel{generic.API, models.Keypresser})
	}
}
