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
	var encodedCommands string
	var encodedConfig string
	if len(os.Args) >= 3 {
		mode = os.Args[1]
		encodedCommands = os.Args[2]
	}
	if len(os.Args) >= 4 {
		encodedConfig = os.Args[3]
	}
	switch mode {
	case "parent":
		// parent is a test mode for spawning a child proc directly from session 0 CLI. see README.md for instructions.
		if err := subproc.SpawnSelf(" child " + encodedCommands + " " + encodedConfig); err != nil {
			panic(err)
		}
	case "child":
		// child is the subprocess started in session 1, by a session 0 parent. it interacts with the user desktop.
		jsonCmd, err := base64.StdEncoding.DecodeString(encodedCommands)
		if err != nil {
			panic(err)
		}
		jsonCfg, err := base64.StdEncoding.DecodeString(encodedConfig)
		if err != nil {
			panic(err)
		}

		logger.Debug("executing keypresses in a child process")
		ctx := context.WithValue(context.Background(), subproc.Flag_InChild, true)
		if err := models.ExecuteJSONEvents(ctx, logger, jsonCmd, jsonCfg); err != nil {
			panic(err)
		}
	default:
		// ModularMain can take multiple APIModel arguments, if your module implements multiple models.
		module.ModularMain(resource.APIModel{API: generic.API, Model: models.Keypresser})
	}
}
