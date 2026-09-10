# Model devrel:generic-sequence-sensor:generic-sequence-sensor

A sensor that stores a set of named sequences — each a list of (resource, method) pairs — and manages capture-frequency overrides at runtime. Start a sequence to activate high-frequency capture on configured resources; stop it to zero out capture and clear the tag.

**API:** `rdk:component:sensor`

## Configuration

The following attribute template can be used to configure this model:

```json
{
  "sequences": [
    {
      "resources": [
        {
          "resource_name": <string>,
          "method": <string>,
          "sequence_cap_hz": <float>,
          "tags": [<string>]
        }
      ]
    }
  ]
}
```

### Attributes

| Name        | Type  | Inclusion | Description                                                          |
| ----------- | ----- | --------- | -------------------------------------------------------------------- |
| `sequences` | array | Required  | One or more sequence definitions. Each must have a `resources` list. |

Each entry in `resources`:

| Name              | Type     | Inclusion | Description                                                                   |
| ----------------- | -------- | --------- | ----------------------------------------------------------------------------- |
| `resource_name`   | string   | Required  | Name of the resource involved in this step.                                   |
| `method`          | string   | Required  | Data-capture method to override. One of `Readings`, `GetImages`, `JointPositions`, `EndPosition`, `DoCommand`. |
| `sequence_cap_hz` | float    | Optional  | Capture frequency (Hz) to apply when the sequence is active. Defaults to `0`. |
| `tags`            | []string | Optional  | Data-capture tags to include in overrides for this resource.                  |

### Example Configuration

```json
{
  "sequences": [
    {
      "resources": [
        {
          "resource_name": "camera-1",
          "method": "GetImages",
          "sequence_cap_hz": 10,
          "tags": ["foo"]
        },
        {
          "resource_name": "arm-1",
          "method": "JointPositions",
          "sequence_cap_hz": 5,
          "tags": ["bar"]
        },
        {
          "resource_name": "gripper-1",
          "method": "DoCommand",
          "sequence_cap_hz": 5
        }
      ]
    }
  ]
}
```

### Using `DoCommand` as a capture method

`DoCommand` is the usual way to record state from resources like grippers that expose it only through `DoCommand`. Unlike the other methods, it cannot be enabled by this sensor's override alone. The RDK's DoCommand collector reads its payload from a `docommand_input` key in the capture method's `additional_params`, so the target resource must have a static data capture entry that supplies it. Set `capture_frequency_hz` to `0` there so the collector stays off until a sequence starts.

```json
{
  "name": "gripper-1",
  "api": "rdk:component:gripper",
  "service_configs": [
    {
      "type": "data_manager",
      "attributes": {
        "capture_methods": [
          {
            "method": "DoCommand",
            "capture_frequency_hz": 0,
            "additional_params": {
              "docommand_input": { "command": "get_state" }
            }
          }
        ]
      }
    }
  ]
}
```

When the sequence starts, this sensor's override supplies the frequency and tags, and the collector sends `docommand_input` to the resource on every tick and records the response.

## Readings

While a sequence is active, returns all configured sequences annotated with the current sequence tag, plus a flat `overrides` list for every resource. `capture_frequency_hz` is the configured `sequence_cap_hz` value.

```json
{
  "sequences": [
    {
      "sequence_tags": ["my-tag"],
      "resources": [
        {"resource_name": "camera-1", "method": "GetImages"},
        {"resource_name": "arm-1",    "method": "JointPositions"}
      ]
    }
  ],
  "overrides": [
    {
      "resource_name": "camera-1",
      "method": "GetImages",
      "capture_frequency_hz": 10,
      "tags": ["foo"]
    },
    {
      "resource_name": "arm-1",
      "method": "JointPositions",
      "capture_frequency_hz": 5,
      "tags": ["bar"]
    }
  ]
}
```

When no sequence is active (stopped, or before any sequence has been started), returns:

```json
{"active": false}
```

## Status

Reports the sensor's own state, not its readings. Available via the standard `GetStatus` RPC on any resource.

```json
{"active": true, "sequence_tag": "my-tag"}
```

## DoCommand

### `start`

Activate the sequence and set the capture-frequency overrides to their configured values.

```json
{"command": "start", "sequence_tag": "my-tag"}
```

Returns `{}`.

### `stop`

Deactivate the sequence. Clears the tag and causes `Readings` to return `{"active": false}`.

```json
{"command": "stop"}
```

Returns `{}`.
