# Keystrokes
### A module to send keystrokes to a machine

## Description
Currently only works for Windows

Send keystrokes to your machine. You can send sequential keystrokes, or press keys simultaneously. Each keystroke is separated by a 100ms gap.

## Atributes
n/a

## Usage
Send a `doCommand` to the component. The structure of the command must be as follows:

```json
{
    "keystrokes": [
        {"type": "<sequential|simultaneous>", "keys": ["<keys>", "<to>", "<press>"]}
    ]
}
```

* `type`: This is the type of keypress, either `sequential` or `simultaneous`. 

    `sequential` keypresses will press and release each key in order. For example, if `"keys": ["hello", " ", "world"]`, the module will first press and release the `h` key, followed by pressing and releasing the `e` key, etc.

    `simultaneous` keypresses will press each key in order, then release each key in reverse order. For example, if `"keys": ["VK_SHIFT", "1"]`, the module will press `SHIFT`, then press `1`, then release `1`, and finally release `SHIFT`. 

* `keys`: An array of keys to press. You can press special keys as well. All special keys being with the prefix `VK_`. The keymap can be found at [`./models/keymap.go`](./models/keymap.go). **NOTE** all special keys (`VK_SHIFT`, `VK_ALT`, etc.), must be their own elements in the `keys` array. If you would like to type out the names of those special keys, you should separate them in the array: `["VK", "_ALT"]`. 

## Example
A `doCommand` with the following command would press the Start button, type in `notepad`, open the first selection (most likely the application "Notepad"), then type `Hello World`, and finally punctuate it with `!`.

```json
{
	"keystrokes": [
		{"type": "simultaneous", "keys": ["VK_META"]},
		{"type": "sequential", "keys": ["notepad", "VK_ENTER"]},
		{"type": "sequential", "keys": ["Hello", " ", "World"]},
		{"type": "simultaneous", "keys": ["VK_SHIFT", "1"]}
	]
}
```
