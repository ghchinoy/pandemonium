package entity_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ghchinoy/pandemonium/pkg/component"
	"github.com/ghchinoy/pandemonium/pkg/entity"
	"github.com/ghchinoy/pandemonium/pkg/types"
)

func TestBaseEntity_ActAndObserve(t *testing.T) {
	ctx := t.Context()
	bus := types.NewEventBus()
	defer bus.Close()
	events := bus.Subscribe()

	recComp := &recordingComponent{name: "ContextA", preActOut: "context-a"}
	actComp := &stubActingComponent{action: "wave hello"}

	ent := &entity.BaseEntity{
		EntityName:   "Alice",
		ActComponent: actComp,
		Components: map[string]component.ContextComponent{
			"ContextA": recComp,
		},
		Bus: bus,
	}

	if got := ent.Name(); got != "Alice" {
		t.Errorf("ent.Name() = %q, want %q", got, "Alice")
	}

	obs := "Bob enters the room."
	if err := ent.Observe(ctx, obs); err != nil {
		t.Fatalf("ent.Observe(ctx, %q) returned unexpected error: %v", obs, err)
	}
	if diff := cmp.Diff([]string{obs}, recComp.observed); diff != "" {
		t.Errorf("ent.Observe(ctx, %q) recorded observations diff (-want +got):\n%s", obs, diff)
	}

	spec := types.ActionSpec{CallToAction: "What does Alice do?", OutputType: types.OutputFree}
	gotAction, err := ent.Act(ctx, spec)
	if err != nil {
		t.Fatalf("ent.Act(ctx, %+v) returned unexpected error: %v", spec, err)
	}
	if gotAction != "wave hello" {
		t.Errorf("ent.Act(ctx, %+v) = %q, want %q", spec, gotAction, "wave hello")
	}
	if diff := cmp.Diff([]string{"wave hello"}, recComp.postActs); diff != "" {
		t.Errorf("ent.Act(ctx, %+v) postActs diff (-want +got):\n%s", spec, diff)
	}
	if recComp.updates != 2 {
		t.Errorf("recComp.updates = %d, want 2 (1 from Observe + 1 from Act)", recComp.updates)
	}

	select {
	case ev := <-events:
		if ev.Type != types.EventThought || ev.Actor != "Alice" {
			t.Errorf("bus event = %+v, want Type=%q Actor=%q", ev, types.EventThought, "Alice")
		}
	default:
		t.Errorf("ent.Act(ctx, %+v) did not publish EventThought to bus", spec)
	}
}

func TestBaseEntity_ComponentErrorPropagation(t *testing.T) {
	ctx := t.Context()
	errPreAct := errors.New("pre-act failed")

	ent := &entity.BaseEntity{
		EntityName:   "Alice",
		ActComponent: &stubActingComponent{action: "noop"},
		Components: map[string]component.ContextComponent{
			"Failing": &recordingComponent{name: "Failing", preActErr: errPreAct},
		},
	}

	spec := types.ActionSpec{CallToAction: "Act"}
	_, err := ent.Act(ctx, spec)
	if !errors.Is(err, errPreAct) {
		t.Errorf("ent.Act(ctx, %+v) error = %v, want %v", spec, err, errPreAct)
	}
}

type recordingComponent struct {
	name      string
	preActOut string
	preActErr error
	observed  []string
	postActs  []string
	updates   int
}

func (r *recordingComponent) Name() string { return r.name }

func (r *recordingComponent) PreAct(_ context.Context, _ types.ActionSpec) (string, error) {
	return r.preActOut, r.preActErr
}

func (r *recordingComponent) PostAct(_ context.Context, actionAttempt string) (string, error) {
	r.postActs = append(r.postActs, actionAttempt)
	return "", nil
}

func (r *recordingComponent) PreObserve(_ context.Context, observation string) (string, error) {
	r.observed = append(r.observed, observation)
	return "", nil
}

func (r *recordingComponent) PostObserve(_ context.Context) (string, error) {
	return "", nil
}

func (r *recordingComponent) Update(_ context.Context) error {
	r.updates++
	return nil
}

type stubActingComponent struct {
	action string
}

func (s *stubActingComponent) Name() string { return "StubAct" }

func (s *stubActingComponent) GetActionAttempt(_ context.Context, _ map[string]string, _ types.ActionSpec) (string, error) {
	return s.action, nil
}
