# Keystrokes + Mouse Inputs
### A module to send keystrokes and mouse inputs to a machine

## Description
Currently only works for Windows

Send keystrokes and mouse inputs to your machine. 
You can send sequential keystrokes, or press keys simultaneously. Each keystroke is separated by a 100ms gap.
You can send left, right, or double left clicks.

## Atributes
The following attributes are available for this model:

| Name              | Type     | Inclusion | Description                    |
|-------------------|----------|-----------|--------------------------------|
| `macros`          | object   | Optional  | Pre-configured macros          |

#### Example Configuration
```json
{
  "macros": {
	"hello_world": [
      {
		"command": "mouse_event",
        "type": "left_click",
        "y": 0.999,
        "x": 0.001
      },
      {
        "command": "sleep",
        "ms": 1000
      },
      {
        "command": "keystroke",
        "mode": "sequential",
        "keys": [
          "notepad",
          "VK_ENTER"
        ]
      },
      {
        "command": "keystroke",
        "mode": "sequential",
        "keys": [
          "Hello",
          " ",
          "World"
        ]
      },
      {
        "command": "keystroke",
        "mode": "simultaneous",
        "keys": [
          "VK_SHIFT",
          "1"
        ]
      }
    ]
  }
}
```

### Macros

This optional attribute allows you to pre-define inputs. After you have defined the inputs, you can call them from the `doCommand`.
From the example configuration above, you could run the `hello_world` macro by calling `doCommand` with the following input:

```json
{
	"inputs": [
		{"command": "macro", "name": "hello_world"}
	]
}
```

## Usage
Send a `doCommand` to the component. The structure of the command must be as follows:

```json
{
    "inputs": [
        {"command": "keystroke", "mode": "<sequential|simultaneous>", "keys": ["<keys>", "<to>", "<press>"]},
		{"command": "mouse_event", "type": "<left_click|right_click|double_click>", "x": <x_coord>, "y": <y_coord>},
		{"command": "sleep", "ms": <MS_TO_SLEEP>},
		{"command": "macro", "name" "<macro_name>"}
    ]
}
```

### Keystrokes
* `type`: This is the type of keypress, either `sequential` or `simultaneous`. 

    `sequential` keypresses will press and release each key in order. For example, if `"keys": ["hello", " ", "world"]`, the module will first press and release the `h` key, followed by pressing and releasing the `e` key, etc.

    `simultaneous` keypresses will press each key in order, then release each key in reverse order. For example, if `"keys": ["VK_SHIFT", "1"]`, the module will press `SHIFT`, then press `1`, then release `1`, and finally release `SHIFT`. 

* `keys`: An array of keys to press. You can press special keys as well. All special keys being with the prefix `VK_`. The keymap can be found at [`./models/keymap.go`](./models/keymap.go). **NOTE** all special keys (`VK_SHIFT`, `VK_ALT`, etc.), must be their own elements in the `keys` array. If you would like to type out the names of those special keys, you should separate them in the array: `["VK", "_ALT"]`. 

### Mouse Inputs
* `type`: This is the type of mouse input, one of `left_click`, `right_click`, or `double_click`.
* `x` and `y`: These comprise the point where the mouse input should occur. These values must be floats between 0 and 1 (inclusive), where (0, 0) is the top left and (1, 1) is the bottom right. These values represent percentages of the view size. For example, if you wish to click directly in the middle of the screen, the `(x, y)` coordinate should be `(0.5, 0.5)`.

	Another example: let's say you have a screenshot with the dimensions of `(600, 600)`. The screenshot is of a computer with a display of dimensions `(1600, 900)`. You click the screenshot at the coordinate `(100, 200)`, and would like that click to be transmitted to the actual machine. You would then make the input coordinate `(0.1666, 0.3333)`, which would translate to the point `(267, 300)` on the actual machine.

## Example
A `doCommand` with the following command would click the Start button (bottom left corner), type in `notepad`, open the first selection (most likely the application "Notepad"), then type `Hello World`, and finally punctuate it with `!`.

```json
{
	"inputs": [
		{"command": "mouse_event", "type": "left_click", "x": 0.001, "y": 0.999},
		{"command": "keystroke", "mode": "sequential", "keys": ["notepad", "VK_ENTER"]},
		{"command": "keystroke", "mode": "sequential", "keys": ["Hello", " ", "World"]},
		{"command": "keystroke", "mode": "simultaneous", "keys": ["VK_SHIFT", "1"]}
	]
}
```
