package genericsequencesensor

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"

	sensor "go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
)

var (
	GenericSequenceSensor = resource.NewModel("devrel", "generic-sequence-sensor", "generic-sequence-sensor")

	// Data-capture method names as the RDK spells them (see components/*/collectors.go).
	validMethods = []string{"Readings", "GetImages", "JointPositions", "EndPosition", "DoCommand"}
)

func init() {
	resource.RegisterComponent(sensor.API, GenericSequenceSensor,
		resource.Registration[sensor.Sensor, *Config]{
			Constructor: newGenericSequenceSensor,
		},
	)
}

type ResourceConfig struct {
	ResourceName  string   `json:"resource_name"`
	Method        string   `json:"method"`
	SequenceCapHz float64  `json:"sequence_cap_hz"`
	Tags          []string `json:"tags"`
}

type SequenceConfig struct {
	Resources []ResourceConfig `json:"resources"`
}

type Config struct {
	Sequences []SequenceConfig `json:"sequences"`
}

func (cfg *Config) Validate(path string) ([]string, []string, error) {
	for i, seq := range cfg.Sequences {
		for j, res := range seq.Resources {
			if res.ResourceName == "" {
				return nil, nil, fmt.Errorf("%s.sequences[%d].resources[%d]: resource_name must not be empty", path, i, j)
			}
			if !slices.Contains(validMethods, res.Method) {
				return nil, nil, fmt.Errorf("%s.sequences[%d].resources[%d]: method %q must be one of %s", path, i, j, res.Method, strings.Join(validMethods, ", "))
			}
		}
	}
	return nil, nil, nil
}

type genericSequenceSensor struct {
	resource.AlwaysRebuild
	resource.Named
	resource.TriviallyCloseable

	logger logging.Logger
	cfg    *Config

	mu             sync.Mutex
	sequenceActive bool
	sequenceTag    string
}

func newGenericSequenceSensor(ctx context.Context, deps resource.Dependencies, rawConf resource.Config, logger logging.Logger) (sensor.Sensor, error) {
	conf, err := resource.NativeConfig[*Config](rawConf)
	if err != nil {
		return nil, err
	}
	return &genericSequenceSensor{
		Named:  rawConf.ResourceName().AsNamed(),
		logger: logger,
		cfg:    conf,
	}, nil
}

func (s *genericSequenceSensor) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	s.mu.Lock()
	active, tag := s.sequenceActive, s.sequenceTag
	s.mu.Unlock()

	if !active {
		return map[string]interface{}{"active": false}, nil
	}

	tags := []interface{}{}
	if tag != "" {
		tags = append(tags, tag)
	}

	sequences := make([]interface{}, len(s.cfg.Sequences))
	overrides := []interface{}{}
	for i, seq := range s.cfg.Sequences {
		resources := make([]interface{}, len(seq.Resources))
		for j, res := range seq.Resources {
			resources[j] = map[string]interface{}{
				"resource_name": res.ResourceName,
				"method":        res.Method,
			}
			overrides = append(overrides, map[string]interface{}{
				"resource_name":        res.ResourceName,
				"method":               res.Method,
				"capture_frequency_hz": res.SequenceCapHz,
				"tags":                 res.Tags,
			})
		}
		sequences[i] = map[string]interface{}{
			"sequence_tags": tags,
			"resources":     resources,
		}
	}

	return map[string]interface{}{
		"sequences": sequences,
		"overrides": overrides,
	}, nil
}

func (s *genericSequenceSensor) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	command, ok := cmd["command"].(string)
	if !ok {
		return nil, fmt.Errorf("command field must be a string")
	}

	switch command {
	case "start":
		tag, ok := cmd["sequence_tag"].(string)
		if !ok {
			return nil, fmt.Errorf("start command requires a sequence_tag string")
		}
		s.mu.Lock()
		s.sequenceActive, s.sequenceTag = true, tag
		s.mu.Unlock()
		return map[string]interface{}{}, nil

	case "stop":
		s.mu.Lock()
		s.sequenceActive, s.sequenceTag = false, ""
		s.mu.Unlock()
		return map[string]interface{}{}, nil

	default:
		return nil, fmt.Errorf("unknown command: %q", command)
	}
}

// Status reports the sensor's own state (whether a sequence is active and its tag), not its readings.
func (s *genericSequenceSensor) Status(ctx context.Context) (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return map[string]interface{}{
		"active":       s.sequenceActive,
		"sequence_tag": s.sequenceTag,
	}, nil
}
