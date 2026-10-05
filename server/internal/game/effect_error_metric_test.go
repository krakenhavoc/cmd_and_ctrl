package game

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics/metricstest"
)

// ADR 0123 §3: every EventEffectError the engine emits moves
// cmdctrl_effect_errors_total by one, at EmitEvent, the one place they
// all pass. A clean stop (ErrStopResolution) is not an error and is
// not emitted, so it is not counted either.
func TestEffectErrorsAreCounted(t *testing.T) {
	g := NewGame()
	count := func() float64 { return metricstest.Value(t, "cmdctrl_effect_errors_total", nil) }

	before := count()
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventEffectError, ErrorMsg: "boom"})
	})
	if got := count() - before; got != 1 {
		t.Errorf("an effect error moved the counter by %v, want 1", got)
	}

	before = count()
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventEffectError, ErrorMsg: "wrapped: " + ErrStopResolution.Error()})
		g.EmitEvent(Event{Kind: EventChangeLife})
	})
	if got := count() - before; got != 0 {
		t.Errorf("a clean stop and an ordinary event moved the counter by %v, want 0", got)
	}
}
