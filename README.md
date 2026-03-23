# generic-gantry

Single-axis and multi-axis gantry components for [Viam](https://viam.com).

## Model viam:generic-gantry:single-axis

A motor-driven linear axis with optional limit switches and encoder homing.

The single-axis gantry converts between motor revolutions and linear millimeters using the `mm_per_rev` configuration. It supports three homing modes depending on how many limit switches are configured:

- **No limit switches**: Homes using the motor encoder only.
- **One limit switch**: Homes to the switch, then uses `mm_per_rev` to calculate the opposite end.
- **Two limit switches**: Homes by driving to each switch to measure the full travel range.

A background goroutine polls limit switches at 1ms intervals and stops the motor immediately on contact, then reverses slightly to move off the switch.

### Attributes

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `motor` | string | **yes** | Name of the motor component driving this axis. |
| `length_mm` | float | **yes** | Total travel length of the axis in millimeters. Must be positive. |
| `mm_per_rev` | float | **yes** | Linear travel per motor revolution in millimeters. Must be positive. |
| `board` | string | no | Name of the board component with limit switch GPIO pins. Required if `limit_pins` is set. |
| `limit_pins` | []string | no | GPIO pin names for limit switches (0, 1, or 2 pins). Requires `board`. |
| `limit_pin_enabled_high` | bool | no | Set `true` if limit switches read high when triggered, `false` if low. Required if `limit_pins` is set. |
| `gantry_mm_per_sec` | float | no | Default travel speed in mm/s when none is specified in a move command. |
| `kinematics_file` | string | no | Path to a kinematics JSON file for motion planning integration. |

### Example configuration

```json
{
  "model": "viam:generic-gantry:single-axis",
  "name": "x-axis",
  "api": "rdk:component:gantry",
  "attributes": {
    "motor": "x-motor",
    "board": "main-board",
    "length_mm": 500,
    "mm_per_rev": 40,
    "limit_pins": ["15", "16"],
    "limit_pin_enabled_high": true,
    "gantry_mm_per_sec": 50
  },
  "depends_on": ["x-motor", "main-board"]
}
```

## Model viam:generic-gantry:multi-axis

Composes multiple single-axis gantries into a coordinated multi-axis system.

Axes are ordered — positions and lengths are returned in the same order as `subaxes_list`. By default axes move sequentially (one at a time in order). Set `move_simultaneously` to `true` to move all axes in parallel.

### Attributes

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `subaxes_list` | []string | **yes** | Ordered list of single-axis gantry component names. At least one is required. |
| `move_simultaneously` | bool | no | If `true`, all axes move at the same time. Defaults to `false` (sequential movement). |

### Example configuration

```json
{
  "model": "viam:generic-gantry:multi-axis",
  "name": "xyz-gantry",
  "api": "rdk:component:gantry",
  "attributes": {
    "subaxes_list": ["x-axis", "y-axis", "z-axis"],
    "move_simultaneously": true
  },
  "depends_on": ["x-axis", "y-axis", "z-axis"]
}
```
