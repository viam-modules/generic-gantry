# generic-gantry

A [Viam](https://viam.com) module providing single-axis and multi-axis gantry components (`rdk:component:gantry`).

## Models

### `viam:generic-gantry:single-axis`

Motor-driven linear axis with optional limit switches and encoder homing.

#### Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `motor` | string | **Yes** | Name of the motor component. Must support position reporting. |
| `length_mm` | float64 | **Yes** | Total travel distance in millimeters. Must be positive. |
| `mm_per_rev` | float64 | **Yes** | Millimeters of linear travel per motor revolution. Must be positive. |
| `board` | string | No | Name of the board component. Required when `limit_pins` is set. |
| `limit_pins` | []string | No | GPIO pin names for limit switches (0, 1, or 2 pins). |
| `limit_pin_enabled_high` | bool | No* | `true` if limit switches read HIGH when triggered, `false` if LOW. *Required when `limit_pins` is set. |
| `gantry_mm_per_sec` | float64 | No | Default movement speed in mm/s. If unset, defaults to 100 motor RPM. |
| `kinematics_file` | string | No | Path to a JSON kinematics model file for frame system integration. |

#### Homing Modes

The number of limit switches determines the homing behavior:

- **0 pins (encoder-only):** Uses the current motor position as zero. Position limits are derived from `length_mm` and `mm_per_rev`.
- **1 pin:** Homes to the single limit switch, then calculates the opposite limit from `length_mm` and `mm_per_rev`.
- **2 pins:** Homes to both the min and max limit switches to establish the full travel range.

#### Example Config

```json
{
  "name": "x-axis",
  "api": "rdk:component:gantry",
  "model": "viam:generic-gantry:single-axis",
  "attributes": {
    "motor": "x-motor",
    "board": "pi",
    "length_mm": 500,
    "mm_per_rev": 40,
    "limit_pins": ["16", "17"],
    "limit_pin_enabled_high": true,
    "gantry_mm_per_sec": 50
  },
  "depends_on": ["x-motor", "pi"]
}
```

---

### `viam:generic-gantry:multi-axis`

Composes multiple single-axis gantries into a multi-axis system.

#### Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `subaxes_list` | []string | **Yes** | Ordered list of single-axis gantry component names. |
| `move_simultaneously` | bool | No | If `true`, all axes move in parallel. If `false` (default), axes move sequentially in order. |

#### Example Config

```json
{
  "name": "xyz-gantry",
  "api": "rdk:component:gantry",
  "model": "viam:generic-gantry:multi-axis",
  "attributes": {
    "subaxes_list": ["x-axis", "y-axis", "z-axis"],
    "move_simultaneously": true
  },
  "depends_on": ["x-axis", "y-axis", "z-axis"]
}
```
