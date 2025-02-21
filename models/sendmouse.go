package models

import (
	"time"
	"unsafe"
)

const flag_MouseInput = 0
const flag_AbsolutePosition = 0x8000
const flag_LeftDown = 0x0002
const flag_LeftUp = 0x0004
const flag_RightDown = 0x0008
const flag_RightUp = 0x0010
const screenNormalizingCoefficient = 65535 // https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-mouse_event#remarks

type mouseInput struct {
	dx          int32
	dy          int32
	mouseData   uint32
	dwFlags     uint32
	time        uint32
	dwExtraInfo uint64
}

type minput struct {
	inputType uint32
	mi        mouseInput
}

// Inputs received from MouseEvents should be floats between 0 and 1,
// representing the distance from the top-left corner of the screen
// as a percentage of the view's width and height.
// This function normalizes these coordinates to the screen's resolution.
func normalizeCoordinates(x, y float64) (dx, dy int32) {
	dx = int32(x * screenNormalizingCoefficient)
	dy = int32(y * screenNormalizingCoefficient)
	return dx, dy
}

func LeftClick(x, y float64) error {
	dx, dy := normalizeCoordinates(x, y)

	i := minput{inputType: flag_MouseInput, mi: mouseInput{dwFlags: flag_LeftDown | flag_AbsolutePosition, dx: dx, dy: dy}}
	if ret, _, err := sendInputProc.Call(
		uintptr(1),
		uintptr(unsafe.Pointer(&i)),
		uintptr(unsafe.Sizeof(i)),
	); ret == 0 {
		return err
	}

	time.Sleep(50 * time.Millisecond)

	i = minput{inputType: flag_MouseInput, mi: mouseInput{dwFlags: flag_LeftUp | flag_AbsolutePosition, dx: dx, dy: dy}}
	if ret, _, err := sendInputProc.Call(
		uintptr(1),
		uintptr(unsafe.Pointer(&i)),
		uintptr(unsafe.Sizeof(i)),
	); ret == 0 {
		return err
	}
	return nil
}

func DoubleClick(x, y float64) error {
	if err := LeftClick(x, y); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)
	if err := LeftClick(x, y); err != nil {
		return err
	}
	return nil
}

func RightClick(x, y float64) error {
	dx, dy := normalizeCoordinates(x, y)

	i := minput{inputType: flag_MouseInput, mi: mouseInput{dwFlags: flag_RightDown | flag_AbsolutePosition, dx: dx, dy: dy}}
	if ret, _, err := sendInputProc.Call(
		uintptr(1),
		uintptr(unsafe.Pointer(&i)),
		uintptr(unsafe.Sizeof(i)),
	); ret == 0 {
		return err
	}

	time.Sleep(50 * time.Millisecond)

	i = minput{inputType: flag_MouseInput, mi: mouseInput{dwFlags: flag_RightUp | flag_AbsolutePosition, dx: dx, dy: dy}}
	if ret, _, err := sendInputProc.Call(
		uintptr(1),
		uintptr(unsafe.Pointer(&i)),
		uintptr(unsafe.Sizeof(i)),
	); ret == 0 {
		return err
	}
	return nil
}
