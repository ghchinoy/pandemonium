package types_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ghchinoy/pandemonium/pkg/types"
)

func TestEventBus_PublishSubscribe(t *testing.T) {
	bus := types.NewEventBus()
	defer bus.Close()

	sub1 := bus.Subscribe()
	sub2 := bus.Subscribe()

	want := types.Event{
		Type:    types.EventAction,
		Actor:   "Alice",
		Message: "shares the cookie",
	}
	bus.Publish(want)

	for i, sub := range []chan types.Event{sub1, sub2} {
		select {
		case got := <-sub:
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("subscriber[%d] received unexpected event diff (-want +got):\n%s", i, diff)
			}
		default:
			t.Errorf("subscriber[%d] did not receive published event %+v", i, want)
		}
	}
}

func TestEventBus_Unsubscribe(t *testing.T) {
	bus := types.NewEventBus()
	defer bus.Close()

	sub1 := bus.Subscribe()
	sub2 := bus.Subscribe()

	bus.Unsubscribe(sub1)

	if _, ok := <-sub1; ok {
		t.Errorf("Unsubscribe(sub1) left channel open, want closed channel")
	}

	want := types.Event{Type: types.EventObservation, Actor: "Bob", Message: "hears the wind"}
	bus.Publish(want)

	select {
	case got := <-sub2:
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("sub2 received unexpected event diff (-want +got):\n%s", diff)
		}
	default:
		t.Errorf("sub2 did not receive event %+v after Unsubscribe(sub1)", want)
	}
}

func TestEventBus_Close(t *testing.T) {
	bus := types.NewEventBus()
	sub := bus.Subscribe()

	bus.Close()
	bus.Close() // Verify idempotency

	if _, ok := <-sub; ok {
		t.Errorf("bus.Close() left subscriber channel open, want closed")
	}

	// Publishing after Close should not panic
	bus.Publish(types.Event{Type: types.EventSimComplete, Message: "done"})

	lateSub := bus.Subscribe()
	if _, ok := <-lateSub; ok {
		t.Errorf("bus.Subscribe() after Close() returned open channel, want closed")
	}
}
