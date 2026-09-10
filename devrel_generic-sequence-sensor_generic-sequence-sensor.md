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
| `method`          | string   | Required  | Method to associate. Must be `Readings`, `GetImages`, or `JointPositions`.    |
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
        }
      ]
    }
  ]
}
```

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
