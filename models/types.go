package models

type EventType string

const (
	Type_Keystroke  EventType = "keystroke"
	Type_MouseEvent EventType = "mouse_event"
	Type_Sleep      EventType = "sleep"
	Type_Macro      EventType = "macro"
)

type Event struct {
	Type EventType `json:"type"`
	Keystroke
	MouseEvent
	Sleep
	Macro
}

type KeypressMode string

const (
	Sequential   KeypressMode = "sequential"
	Simultaneous KeypressMode = "simultaneous"
)

type Keystroke struct {
	Mode KeypressMode `json:"mode"`
	Keys []string     `json:"keys"`
}

type MouseEventType string

const (
	EventLeftClick   MouseEventType = "left_click"
	EventRightClick  MouseEventType = "right_click"
	EventDoubleClick MouseEventType = "double_click"
)

type MouseEvent struct {
	Event MouseEventType `json:"event"`
	X     float64        `json:"x"`
	Y     float64        `json:"y"`
}

type Sleep struct {
	Ms int `json:"ms"`
}

type Macro struct {
	Name   string  `json:"name"`
	Events []Event `json:"events"`
}
