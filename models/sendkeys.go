package models

import (
	"syscall"
	"time"
	"unsafe"
)

var user32 = syscall.NewLazyDLL("user32.dll")
var sendInputProc = user32.NewProc("SendInput")

type keyboardInput struct {
	wVk         uint16
	wScan       uint16
	dwFlags     uint32
	time        uint32
	dwExtraInfo uint64
}

type input struct {
	inputType uint32
	ki        keyboardInput
	mi        mouseInput
	padding   uint64
}

const flag_KeyboardInput = 1
const flag_KeyUp = 0x0002

func keyWithShift(key int) (int, bool) {
	if key < 0xFFF {
		return key, false
	}
	return key - 0xFFF, true
}

func keyDown(key int) error {
	input := input{inputType: flag_KeyboardInput, ki: keyboardInput{wVk: uint16(key)}}
	if ret, _, err := sendInputProc.Call(
		uintptr(1),
		uintptr(unsafe.Pointer(&input)),
		uintptr(unsafe.Sizeof(input)),
	); ret == 0 {
		return err
	}
	return nil
}

func keyUp(key int) error {
	input := input{inputType: flag_KeyboardInput, ki: keyboardInput{wVk: uint16(key), dwFlags: flag_KeyUp}}
	if ret, _, err := sendInputProc.Call(
		uintptr(1),
		uintptr(unsafe.Pointer(&input)),
		uintptr(unsafe.Sizeof(input)),
	); ret == 0 {
		return err
	}
	return nil
}

func Press(key int) error {
	key, needsShift := keyWithShift(key)
	if needsShift {
		if err := keyDown(keymap["VK_SHIFT"]); err != nil {
			return err
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err := keyDown(key); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	return nil
}

func Release(key int) error {
	key, needsShift := keyWithShift(key)
	if needsShift {
		defer func() {
			keyUp(keymap["VK_SHIFT"])
		}()
	}
	if err := keyUp(key); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	return nil
}
