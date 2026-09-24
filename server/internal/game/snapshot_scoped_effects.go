package game

import (
	"fmt"
	"strings"
)

// snapshot_scoped_effects.go is the snapshot half of ADR 0041 phase 3's
// data records (#1497): the deep copy at the snapshot boundary, and the
// restore-time refusal of an effect key this binary cannot interpret
// (ErrUnknownEffectKey, ADR 0041 P4). The record itself, its vocabulary
// and its interpreter are in scoped_effects.go.

// ---------------------------------------------------------------
// Restore validation
// ---------------------------------------------------------------

// checkEffectKeys refuses a snapshot naming an effect key this binary
// cannot interpret (ADR 0041 P4): a mod kind, or a delayed-trigger or
// stack-item body or condition key. Every key a binary can write is
// one it registered, so the only way to meet an unknown one is a
// ROLLBACK — the file was written by a newer build — and the answer is
// the one Decision 5 gives a too-new file: refuse it, keep it.
//
// Effect-body keys arrive with ADR 0041 P2 (tier 2). This binary has
// none registered, so it refuses any non-empty one: that is the reader
// the v7 bump promised, in place before the writer exists.
func (s *GameSnapshot) checkEffectKeys() error {
	var unknown []string
	for _, e := range s.ScopedEffects {
		for _, m := range e.Mods {
			if !KnownModKind(m.Kind) {
				unknown = append(unknown, "mod kind "+string(m.Kind))
			}
		}
	}
	for _, d := range s.DelayedTriggers {
		if d.Body != "" {
			unknown = append(unknown, "delayed-trigger body "+d.Body)
		}
		if d.Condition != "" {
			unknown = append(unknown, "delayed-trigger condition "+d.Condition)
		}
	}
	for _, list := range [][]stackItemSnapshot{s.StackMeta, s.PendingTriggers} {
		for _, it := range list {
			if it.Body != "" {
				unknown = append(unknown, "stack-item body "+it.Body)
			}
		}
	}
	if len(unknown) == 0 {
		return nil
	}
	return fmt.Errorf("%w: %s", ErrUnknownEffectKey, strings.Join(unknown, ", "))
}

// deepCopyScopedEffects copies records and their inner slices, for the
// snapshot boundary, where nothing may alias live game state.
func deepCopyScopedEffects(in []ScopedEffect) []ScopedEffect {
	if len(in) == 0 {
		return nil
	}
	out := make([]ScopedEffect, len(in))
	for i, e := range in {
		e.Affected = append([]AffectedObject(nil), e.Affected...)
		e.Mods = cloneMods(e.Mods)
		out[i] = e
	}
	return out
}
