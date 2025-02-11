package models

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/viam-labs/screenshot-cam/subproc"
	"go.viam.com/rdk/components/generic"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"golang.org/x/sys/windows/svc"
)

var (
	Keypresser = resource.NewModel("awinter", "keystrokes", "keypresser")
)

func init() {
	resource.RegisterComponent(generic.API, Keypresser,
		resource.Registration[resource.Resource, *Config]{
			Constructor: newKeystrokesKeypresser,
		},
	)
}

type Config struct {
	DelayMS uint8 `json:"delay_ms"`
	resource.TriviallyValidateConfig
}

type keystrokesKeypresser struct {
	name resource.Name

	logger logging.Logger
	cfg    *Config

	cancelCtx  context.Context
	cancelFunc func()

	resource.TriviallyReconfigurable
	resource.TriviallyCloseable
}

func newKeystrokesKeypresser(ctx context.Context, deps resource.Dependencies, rawConf resource.Config, logger logging.Logger) (resource.Resource, error) {
	conf, err := resource.NativeConfig[*Config](rawConf)
	if err != nil {
		return nil, err
	}

	cancelCtx, cancelFunc := context.WithCancel(context.Background())

	s := &keystrokesKeypresser{
		name:       rawConf.ResourceName(),
		logger:     logger,
		cfg:        conf,
		cancelCtx:  cancelCtx,
		cancelFunc: cancelFunc,
	}
	return s, nil
}

func (s *keystrokesKeypresser) Name() resource.Name {
	return s.name
}

type KeypressType string

const (
	Sequential   KeypressType = "sequential"
	Simultaneous KeypressType = "simultaneous"
)

type Keystroke struct {
	Type KeypressType `json:"type"`
	Keys []string     `json:"keys"`
}
type Command struct {
	Keystrokes []Keystroke `json:"keystrokes"`
}

func (s *keystrokesKeypresser) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	jsonbody, err := json.Marshal(cmd)
	if err != nil {
		return nil, fmt.Errorf("could not convert command into JSON: %w", err)
	}

	if isService, err := svc.IsWindowsService(); err != nil {
		return nil, err
	} else if isService {
		jsonArg := base64.StdEncoding.EncodeToString(jsonbody)
		// parent is a test mode for spawning a child proc directly from session 0 CLI. see README.md for instructions.
		return nil, subproc.SpawnSelf(" child " + jsonArg)
	} else {
		var command Command
		err := json.Unmarshal(jsonbody, &command)
		if err != nil {
			return nil, err
		}
		return nil, doKeystrokes(command)
	}
}

func doKeystrokes(command Command) error {
	for _, keystroke := range command.Keystrokes {
		if keystroke.Type == Simultaneous {
			pressed := []int{}
			for _, keys := range keystroke.Keys {
				// Check if meta key and press/release immediately
				// Otherwise, go rune by rune
				if key, ok := keymap[keys]; ok {
					if err := Press(key); err != nil {
						return err
					}
					pressed = append(pressed, key)
				} else {
					for _, r := range keys {
						if key := GetKey(r); key >= 0 {
							if err := Press(key); err != nil {
								return err
							}
							pressed = append(pressed, key)
						}
					}
				}
			}
			for i := len(pressed) - 1; i >= 0; i-- {
				if err := Release(pressed[i]); err != nil {
					return err
				}
			}
		} else if keystroke.Type == Sequential {
			for _, keys := range keystroke.Keys {
				// Check if meta key and press/release immediately
				// Otherwise, go rune by rune
				if key, ok := keymap[keys]; ok {
					if err := Press(key); err != nil {
						return err
					}
					if err := Release(key); err != nil {
						return err
					}
				} else {
					for _, r := range keys {
						if key := GetKey(r); key >= 0 {
							if err := Press(key); err != nil {
								return err
							}
							if err := Release(key); err != nil {
								return err
							}
						}
					}
				}
			}
		}
	}
	return nil
}

func DemoMode(ctx context.Context, logger logging.Logger, jsonArg []byte) error {
	var command Command
	err := json.Unmarshal(jsonArg, &command)
	if err != nil {
		return err
	}
	return doKeystrokes(command)
}
