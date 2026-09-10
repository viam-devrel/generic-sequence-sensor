# generic-sequence-sensor Module

The `devrel:generic-sequence-sensor` module provides a sensor that stores a set of named sequences — each a list of (resource, method) pairs — and manages capture-frequency overrides at runtime. Start a sequence to activate high-frequency capture on configured resources; stop it to zero out capture and clear the tag.

## Models

- [`devrel:generic-sequence-sensor:generic-sequence-sensor`](devrel_generic-sequence-sensor_generic-sequence-sensor.md) — `rdk:component:sensor`. Configuration, readings, and DoCommand reference.
