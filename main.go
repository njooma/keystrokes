package main

import (
	"context"
	"encoding/base64"
	"keystrokes/models"
	"os"

	"go.viam.com/rdk/components/generic"
	"go.viam.com/rdk/module"
	"go.viam.com/rdk/resource"

	"keystrokes/subproc"
)

func main() {
	logger := module.NewLoggerFromArgs("keystrokes")
	var mode string
	var encodedKeystrokes string
	if len(os.Args) >= 3 {
		mode = os.Args[1]
		encodedKeystrokes = os.Args[2]
	}
	switch mode {
	case "parent":
		// parent is a test mode for spawning a child proc directly from session 0 CLI. see README.md for instructions.
		if err := subproc.SpawnSelf(" child " + encodedKeystrokes); err != nil {
			panic(err)
		}
	case "child":
		// child is the subprocess started in session 1, by a session 0 parent. it interacts with the user desktop.
		jsonArg, err := base64.StdEncoding.DecodeString(encodedKeystrokes)
		if err != nil {
			panic(err)
		}
		logger.Debug("executing keypresses in a child process")
		ctx := context.WithValue(context.Background(), subproc.Flag_InChild, true)
		if err := models.ExecuteJSONEvents(ctx, logger, jsonArg); err != nil {
			panic(err)
		}
	default:
		// ModularMain can take multiple APIModel arguments, if your module implements multiple models.
		module.ModularMain(resource.APIModel{API: generic.API, Model: models.Keypresser})
	}
}
