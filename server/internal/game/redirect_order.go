package game

import (
	"slices"

	"github.com/google/uuid"
)

// redirect_order.go — #2066: the CR 616 order two redirections are applied
// in is chosen once per damage instance.
//
// The Gideon's Sacrifice and Saving Grace rulings: "When two copies (or two
// such effects) redirect simultaneous damage, the player chooses one
// creature, and all of that damage goes to it." The damage is one instance
// (ADR 0108 PR 0) but one event per source and recipient, and the ordering
// prompt belongs to an event, so asking it per event let the player send
// different events of one instance to different creatures. The answer is
// kept for the instance instead, and every later event of it whose
// applicable effects are all redirections the answer ordered takes that
// order without asking.
//
// Only a window made entirely of redirections is remembered or replayed.
// A prevention shield or a doubler beside them is an ordering question the
// rulings say nothing about, and it is still asked per event.

// redirectOrder is the order the affected player chose for the
// redirections that applied to one damage instance. Plain data, copy on
// write, never captured (a restore starts with none: the instance stamp is
// transient, see snapshot_drift_test.go).
type redirectOrder struct {
	instance DamageInstance
	order    []ReplacementEffectID
}

// onlyRedirections reports whether every gathered effect is a redirection.
func onlyRedirections(applicable []activeReplacement) bool {
	for _, a := range applicable {
		if !a.effect.RedirectsDamage {
			return false
		}
	}
	return len(applicable) > 0
}

// rememberRedirectOrderLocked keeps the order the player gave a window of
// redirections for the instance of ev. A window with anything else in it,
// or an event with no instance, is not remembered.
//
// Caller must hold g.mu (write).
func (g *Game) rememberRedirectOrderLocked(ev *ReplacementEvent, applicable []activeReplacement, ordered []ReplacementEffectID) {
	if ev == nil || ev.Kind != RepEventDamage || ev.DamageInstance == 0 || !onlyRedirections(applicable) {
		return
	}
	next := make([]redirectOrder, 0, len(g.redirectOrders)+1)
	for _, r := range g.redirectOrders {
		if r.instance != ev.DamageInstance {
			next = append(next, r)
		}
	}
	g.redirectOrders = append(next, redirectOrder{instance: ev.DamageInstance, order: slices.Clone(ordered)})
}

// rememberedRedirectOrderLocked reorders a window of redirections into the
// order already chosen for ev's instance, and reports false when there is
// none or it does not cover every effect in the window (a redirection that
// is new to the instance is a fresh question).
//
// Caller must hold g.mu.
func (g *Game) rememberedRedirectOrderLocked(ev *ReplacementEvent, applicable []activeReplacement) ([]activeReplacement, bool) {
	if ev.Kind != RepEventDamage || ev.DamageInstance == 0 || !onlyRedirections(applicable) {
		return nil, false
	}
	for _, r := range g.redirectOrders {
		if r.instance != ev.DamageInstance {
			continue
		}
		rank := func(a activeReplacement) int { return slices.Index(r.order, a.id) }
		for _, a := range applicable {
			if rank(a) < 0 {
				return nil, false
			}
		}
		out := slices.Clone(applicable)
		slices.SortStableFunc(out, func(x, y activeReplacement) int { return rank(x) - rank(y) })
		return out, true
	}
	return nil, false
}

// dropRedirectOrdersLocked forgets the order chosen for instance `inst`
// once it has ended — or, with zero, for every instance not still waiting
// on a prompt (as play moves on, beginEventBatchLocked).
//
// Caller must hold g.mu (write).
func (g *Game) dropRedirectOrdersLocked(inst DamageInstance) {
	if len(g.redirectOrders) == 0 {
		return
	}
	var kept []redirectOrder
	for _, r := range g.redirectOrders {
		ended := r.instance == inst
		if inst == 0 {
			ended = !g.damageInstancePausedLocked(r.instance)
		}
		if !ended {
			kept = append(kept, r)
		}
	}
	g.redirectOrders = kept
}

// settleSiblingRedirectOrdersLocked answers, with the order just given, the
// other ordering prompts of the same damage instance that are windows of
// the same redirections. Simultaneous events each reach the CR 616 window
// before the first prompt is answered, so each has queued its own; the
// player chose one order for the instance, so they take it rather than
// being asked again (#2066). A prompt over a different set of effects is
// left for the player.
//
// Caller must hold g.mu (write).
func (g *Game) settleSiblingRedirectOrdersLocked(inst DamageInstance, chooser uuid.UUID, ordered []ReplacementEffectID) {
	if inst == 0 {
		return
	}
	for guard := 0; guard < 64; guard++ {
		var next *PendingChoice
		for _, c := range g.PendingChoices {
			if c == nil || c.Kind != PendingChoiceReplacementOrder {
				continue
			}
			f := c.replacementResume
			if f == nil || f.ev == nil || f.ev.Kind != RepEventDamage || f.ev.DamageInstance != inst ||
				c.Chooser != chooser || !onlyRedirections(f.applicable) ||
				!sameEffectIDSet(c.ReplacementEffectIDs, ordered) {
				continue
			}
			next = c
			break
		}
		if next == nil {
			return
		}
		if err := g.resolveReplacementOrderLocked(next.ID, chooser, ordered); err != nil {
			return
		}
	}
}

// sameEffectIDSet reports whether a and b hold the same IDs.
func sameEffectIDSet(a, b []ReplacementEffectID) bool {
	if len(a) != len(b) {
		return false
	}
	x, y := slices.Clone(a), slices.Clone(b)
	slices.Sort(x)
	slices.Sort(y)
	return slices.Equal(x, y)
}
