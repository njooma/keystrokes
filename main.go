package main

import (
	"context"
	"encoding/base64"
	"keystrokes/models"
	"os"
	"time"

	"github.com/micmonay/keybd_event"
	"github.com/viam-labs/screenshot-cam/subproc"
	"go.viam.com/rdk/components/generic"
	"go.viam.com/rdk/module"
	"go.viam.com/rdk/resource"
)

func SendKeys() {
	time.Sleep(time.Second * 2)
	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		panic(err)
	}
	kb.SetKeys(keybd_event.VK_A, keybd_event.VK_B)
	if err := kb.Launching(); err != nil {
		panic(err)
	}
}

func main() {
	logger := module.NewLoggerFromArgs("screenshot-cam")
	var arg string
	var arg2 string
	if len(os.Args) >= 3 {
		arg = os.Args[1]
		arg2 = os.Args[2]
	}
	switch arg {
	case "parent":
		arg2 := base64.StdEncoding.EncodeToString([]byte(`{"keystrokes": [{"type":"sequential", "keys": ["A", "B", "C"]}]}`))
		// parent is a test mode for spawning a child proc directly from session 0 CLI. see README.md for instructions.
		if err := subproc.SpawnSelf(" child " + arg2); err != nil {
			panic(err)
		}
	case "child":
		jsonArg, err := base64.StdEncoding.DecodeString(arg2)
		if err != nil {
			panic(err)
		}
		// child is the subprocess started in session 1 by a session 0 parent. it does the work.
		logger.Info("child mode: doing a keystrokes test instead of starting module")
		if err := models.DemoMode(context.Background(), logger, jsonArg); err != nil {
			panic(err)
		}
	default:
		// ModularMain can take multiple APIModel arguments, if your module implements multiple models.
		module.ModularMain(resource.APIModel{generic.API, models.Keypresser})
	}
}
