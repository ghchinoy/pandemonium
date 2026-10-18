package component

import (
	"context"
	"errors"
	"maps"
	"sync"

	"github.com/ghchinoy/pandemonium/pkg/types"
)

// ErrInputClosed is returned when a HumanActComponent is closed while waiting for input.
var ErrInputClosed = errors.New("human input closed")

// HumanInputRequest describes a pending action prompt awaiting human input.
type HumanInputRequest struct {
	Contexts map[string]string
	Spec     types.ActionSpec
}

// HumanActComponent implements ActingComponent for human-in-the-loop play
// (ported from Concordia PR #419's HumanActComponent / BlockingInput).
// When GetActionAttempt is called, it publishes a pending HumanInputRequest and
// blocks until Submit(answer), Close(), or ctx cancellation.
type HumanActComponent struct {
	CompName string

	mu      sync.Mutex
	pending *HumanInputRequest
	replyCh chan string
	closed  bool
}

// NewHumanActComponent creates a new HumanActComponent.
func NewHumanActComponent(name string) *HumanActComponent {
	if name == "" {
		name = "HumanAct"
	}
	return &HumanActComponent{CompName: name}
}

// Name returns the name of the acting component.
func (h *HumanActComponent) Name() string {
	if h.CompName != "" {
		return h.CompName
	}
	return "HumanAct"
}

// GetActionAttempt publishes the context and spec as a pending request and waits for Submit.
func (h *HumanActComponent) GetActionAttempt(ctx context.Context, contexts map[string]string, spec types.ActionSpec) (string, error) {
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return "", ErrInputClosed
	}
	replyCh := make(chan string, 1)
	h.pending = &HumanInputRequest{
		Contexts: maps.Clone(contexts),
		Spec:     spec,
	}
	h.replyCh = replyCh
	h.mu.Unlock()

	select {
	case <-ctx.Done():
		h.clearPending(replyCh)
		return "", ctx.Err()
	case answer, ok := <-replyCh:
		if !ok {
			return "", ErrInputClosed
		}
		return answer, nil
	}
}

// Pending returns a copy of the currently waiting HumanInputRequest, if any.
func (h *HumanActComponent) Pending() (HumanInputRequest, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.pending == nil {
		return HumanInputRequest{}, false
	}
	return HumanInputRequest{
		Contexts: maps.Clone(h.pending.Contexts),
		Spec:     h.pending.Spec,
	}, true
}

// Submit delivers a human player's action answer to the waiting GetActionAttempt call.
func (h *HumanActComponent) Submit(answer string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.pending == nil || h.replyCh == nil || h.closed {
		return false
	}
	h.replyCh <- answer
	h.pending = nil
	h.replyCh = nil
	return true
}

// Close unblocks any waiting GetActionAttempt with ErrInputClosed.
func (h *HumanActComponent) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return
	}
	h.closed = true
	if h.replyCh != nil {
		close(h.replyCh)
		h.replyCh = nil
		h.pending = nil
	}
}

func (h *HumanActComponent) clearPending(ch chan string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.replyCh == ch {
		h.pending = nil
		h.replyCh = nil
	}
}
