package genericsequencesensor

import (
	"context"
	"testing"

	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
)

func TestStartStop(t *testing.T) {
	ctx := context.Background()
	cfg := &Config{Sequences: []SequenceConfig{{Resources: []ResourceConfig{
		{ResourceName: "cam", Method: "GetImages", SequenceCapHz: 10, Tags: []string{"foo"}},
	}}}}
	if _, _, err := cfg.Validate("test"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := (&Config{Sequences: []SequenceConfig{{Resources: []ResourceConfig{{ResourceName: "x", Method: "Nope"}}}}}).Validate("test"); err == nil {
		t.Fatal("expected invalid method to fail validation")
	}

	s := &genericSequenceSensor{Named: resource.NewName(resource.APINamespaceRDK.WithComponentType("sensor"), "s").AsNamed(), logger: logging.NewTestLogger(t), cfg: cfg}

	r, err := s.Readings(ctx, nil)
	if err != nil || r["active"] != false {
		t.Fatalf("inactive readings = %v, %v", r, err)
	}
	if _, err := s.DoCommand(ctx, map[string]interface{}{"command": "start", "sequence_tag": "run1"}); err != nil {
		t.Fatal(err)
	}
	r, _ = s.Readings(ctx, nil)
	seq := r["sequences"].([]interface{})[0].(map[string]interface{})
	if tags := seq["sequence_tags"].([]interface{}); len(tags) != 1 || tags[0] != "run1" {
		t.Fatalf("sequence_tags = %v", tags)
	}
	ov := r["overrides"].([]interface{})[0].(map[string]interface{})
	if ov["capture_frequency_hz"] != 10.0 {
		t.Fatalf("override = %v", ov)
	}
	st, _ := s.Status(ctx)
	if st["active"] != true || st["sequence_tag"] != "run1" {
		t.Fatalf("status = %v", st)
	}
	if _, err := s.DoCommand(ctx, map[string]interface{}{"command": "stop"}); err != nil {
		t.Fatal(err)
	}
	st, _ = s.Status(ctx)
	if st["active"] != false || st["sequence_tag"] != "" {
		t.Fatalf("status after stop = %v", st)
	}
	if _, err := s.DoCommand(ctx, map[string]interface{}{"command": "bogus"}); err == nil {
		t.Fatal("expected unknown command error")
	}
}

func TestStatusOverridesEmbeddedNamed(t *testing.T) {
	s := &genericSequenceSensor{Named: resource.NewName(resource.APINamespaceRDK.WithComponentType("sensor"), "s").AsNamed(), cfg: &Config{}}
	var r resource.Resource = s
	s.sequenceActive, s.sequenceTag = true, "x"
	st, err := r.Status(context.Background())
	if err != nil || st["sequence_tag"] != "x" {
		t.Fatalf("resource.Resource.Status = %v, %v; embedded Named default was not overridden", st, err)
	}
}
