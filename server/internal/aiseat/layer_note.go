package aiseat

import (
	"context"
	"sync/atomic"
)

// layer_note.go is how cmdctrl_bot_decisions_total learns which layer
// answered a window (ADR 0123 §3) without the runner asking for a
// Trace, which it only does when an Observer wants one.
//
// The runner puts a note in the decision's context; a policy that
// knows which layer answered writes it with NoteLayer. The last write
// wins, so a wrapper that delegates (rules.Filter over the model
// funnel) and the policy that finally answers can both write without
// counting twice: the runner counts the window once, after Decide
// returns. The heuristic writes nothing, and a window nobody named is
// Layer B, which is the heuristic's layer.

type layerNoteKey struct{}

type layerNote struct{ layer atomic.Value }

// NoteLayer records which layer answered the window ctx belongs to:
// "A", "B", "C" or "random". It does nothing in a context the runner
// did not make, so a policy may call it unconditionally.
func NoteLayer(ctx context.Context, layer string) {
	if n, ok := ctx.Value(layerNoteKey{}).(*layerNote); ok {
		n.layer.Store(layer)
	}
}

// WithLayerNote returns a context that collects NoteLayer, and the
// func that reads what was noted ("" when nothing was). The runner
// uses it for every decision; it is exported for a harness or a test
// that calls a policy directly.
func WithLayerNote(ctx context.Context) (context.Context, func() string) {
	n := &layerNote{}
	return context.WithValue(ctx, layerNoteKey{}, n), func() string {
		s, _ := n.layer.Load().(string)
		return s
	}
}
