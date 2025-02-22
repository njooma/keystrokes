package models

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"go.viam.com/rdk/components/generic"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"

	"keystrokes/subproc"
)

var (
	Keypresser = resource.NewModel("njooma", "keystrokes", "keypresser")
)

func init() {
	resource.RegisterComponent(generic.API, Keypresser,
		resource.Registration[resource.Resource, *Config]{
			Constructor: newKeystrokesKeypresser,
		},
	)
}

type Config struct {
	Macros map[string][]Command `json:"macros"`
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

type doCommand struct {
	Commands []Command `json:"inputs"`
}

func (s *keystrokesKeypresser) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	jsonCmd, err := json.Marshal(cmd)
	if err != nil {
		return nil, fmt.Errorf("could not convert command into JSON: %w", err)
	}

	jsonCfg, err := json.Marshal(s.cfg)
	if err != nil {
		return nil, fmt.Errorf("could not convert config into JSON: %w", err)
	}

	if subproc.ShouldSpawn(ctx) {
		s.logger.Debug("Running in service mode, spawning child process")
		jsonCmdB64 := base64.StdEncoding.EncodeToString(jsonCmd)
		jsonCfgB64 := base64.StdEncoding.EncodeToString(jsonCfg)
		// Spawn a subprocess to run in ChildMode if we are in a Windows service
		return nil, subproc.SpawnSelf(" child " + jsonCmdB64 + " " + jsonCfgB64)
	}
	s.logger.Debug("Running in interactive mode, executing keypresses directly")
	return nil, ExecuteJSONEvents(ctx, s.logger, jsonCmd, jsonCfg)
}

func handleEvents(commands []Command, cfg Config) error {
	for _, event := range commands {
		switch event.Type {
		case Type_Keystroke:
			if err := doKeystroke(event.Keystroke); err != nil {
				return err
			}
		case Type_MouseEvent:
			if err := doMouseEvent(event.MouseEvent); err != nil {
				return err
			}
		case Type_Sleep:
			doSleep(event.Sleep)
		case Type_Macro:
			if macro, ok := cfg.Macros[event.Macro.Name]; ok {
				if err := handleEvents(macro, cfg); err != nil {
					return err
				}
			} else {
				return fmt.Errorf("unknown macro: %s", event.Macro.Name)
			}
		default:
			return fmt.Errorf("unknown event type: %s", event.Type)
		}
	}
	return nil
}

func doKeystroke(keystroke Keystroke) error {
	if keystroke.Mode == Simultaneous {
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
	} else if keystroke.Mode == Sequential {
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
	return nil
}

func doMouseEvent(mouseEvent MouseEvent) error {
	switch mouseEvent.Event {
	case EventLeftClick:
		return LeftClick(mouseEvent.X, mouseEvent.Y)
	case EventRightClick:
		return RightClick(mouseEvent.X, mouseEvent.Y)
	case EventDoubleClick:
		return DoubleClick(mouseEvent.X, mouseEvent.Y)
	}
	return nil
}

func doSleep(sleep Sleep) {
	time.Sleep(time.Duration(sleep.Ms) * time.Millisecond)
}

// Receive a JSON-encoded Command object, which contains a list of Keystroke objects, and execute it.
func ExecuteJSONEvents(ctx context.Context, logger logging.Logger, jsonCmd []byte, jsonCfg []byte) error {
	var command doCommand
	err := json.Unmarshal(jsonCmd, &command)
	if err != nil {
		return err
	}

	var cfg Config
	json.Unmarshal(jsonCfg, &cfg)
	return handleEvents(command.Commands, cfg)
}
