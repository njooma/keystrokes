package models

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os/user"
	"strings"

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

type MouseEventType string

const (
	EventLeftClick   MouseEventType = "left_click"
	EventRightClick  MouseEventType = "right_click"
	EventDoubleClick MouseEventType = "double_click"
)

type MouseEvent struct {
	Type MouseEventType `json:"type"`
	X    float64        `json:"x"`
	Y    float64        `json:"y"`
}

type Command struct {
	Inputs []interface{} `json:"inputs"`
}

func (c *Command) UnmarshalJSON(data []byte) error {
	var items struct {
		Inputs []json.RawMessage `json:"inputs"`
	}
	if err := json.Unmarshal(data, &items); err != nil {
		return fmt.Errorf("unable to parse command: %w", err)
	}

	for _, msg := range items.Inputs {
		var meta struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(msg, &meta); err != nil {
			return fmt.Errorf("unable to parse command: %w", err)
		}
		switch meta.Type {
		case string(Sequential), string(Simultaneous):
			var keystroke Keystroke
			if err := json.Unmarshal(msg, &keystroke); err != nil {
				return fmt.Errorf("unable to parse command: %w", err)
			}
			c.Inputs = append(c.Inputs, keystroke)
		case string(EventLeftClick), string(EventRightClick), string(EventDoubleClick):
			var mouseEvent MouseEvent
			if err := json.Unmarshal(msg, &mouseEvent); err != nil {
				return fmt.Errorf("unable to parse command: %w", err)
			}
			c.Inputs = append(c.Inputs, mouseEvent)
		default:
			return fmt.Errorf("unable to parse command, invalid type: %s", meta.Type)
		}
	}
	return nil
}

func (s *keystrokesKeypresser) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	jsonbody, err := json.Marshal(cmd)
	if err != nil {
		return nil, fmt.Errorf("could not convert command into JSON: %w", err)
	}

	if username, err := user.Current(); err != nil {
		return nil, err
	} else if strings.Contains(username.Username, `NT AUTHORITY\`) {
		s.logger.Debug("Running in service mode, spawning child process")
		jsonArg := base64.StdEncoding.EncodeToString(jsonbody)
		// Spawn a subprocess to run in ChildMode if we are in a Windows service
		return nil, subproc.SpawnSelf(" child " + jsonArg)
	}
	s.logger.Debug("Running in interactive mode, executing keypresses directly")
	return nil, ExecuteJSONEvents(ctx, s.logger, jsonbody)
}

func handleEvents(command Command) error {
	for _, input := range command.Inputs {
		if event, ok := input.(Keystroke); ok {
			if err := doKeystroke(event); err != nil {
				return err
			}
		}
		if event, ok := input.(MouseEvent); ok {
			if err := doMouseEvent(event); err != nil {
				return err
			}
		}
	}
	return nil
}

func doKeystroke(keystroke Keystroke) error {
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
	return nil
}

func doMouseEvent(mouseEvent MouseEvent) error {
	switch mouseEvent.Type {
	case EventLeftClick:
		return LeftClick(mouseEvent.X, mouseEvent.Y)
	case EventRightClick:
		return RightClick(mouseEvent.X, mouseEvent.Y)
	case EventDoubleClick:
		return DoubleClick(mouseEvent.X, mouseEvent.Y)
	}
	return nil
}

// Receive a JSON-encoded Command object, which contains a list of Keystroke objects, and execute it.
func ExecuteJSONEvents(ctx context.Context, logger logging.Logger, jsonArg []byte) error {
	var command Command
	err := command.UnmarshalJSON(jsonArg)
	if err != nil {
		return err
	}
	return handleEvents(command)
}
