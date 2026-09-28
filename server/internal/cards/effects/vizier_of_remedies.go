package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vizier of Remedies — Creature — Human Monk {1}{G}{W}, 2/2 (Modern
// Horizons):
//
//	"If one or more -1/-1 counters would be put on a creature you
//	 control, that many -1/-1 counters minus one are put on it
//	 instead."
//
// Hardened Scales' mirror on the minus side — one counter FEWER
// instead of one more — and the reason #1710 exists. Devoted Druid's
// "Put a -1/-1 counter on this creature: Untap this creature" pays
// that counter as a COST (CR 601.2h / CR 602.2b), not the effect of a
// resolving spell or ability, so CR 614.16's "if AN EFFECT would put"
// gate (Doubling Season) never reaches it. But this replacement names
// no effect at all — "if one or more counters WOULD BE PUT" — which is
// CR 614.16's other case, the one a cost's counters do not escape: it
// reduces the Druid's one counter to zero, and the Druid untaps for
// free, forever. That is the long-standing Devoted Druid + Vizier of
// Remedies ruling, and it is why a cost-paid counter has to open the
// CR 614 window at all rather than writing the counter map directly
// (payCostCounterLocked, game/counter_cost.go; ADR 0073's 2026-09-28
// amendment).
//
// The predicate gates on a positive delta before this ever runs, so
// CounterDelta-- lands at zero and never below it.
func init() {
	Register(Spec{
		OracleID:     "79770e65-740a-44c7-bea2-a24e6a722c22",
		Name:         "Vizier of Remedies",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			{
				Watches:   []game.EventKind{game.EventCounterPlaced},
				AppliesTo: minusOneCounterPlacementOnYourCreature,
				Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
					ev.CounterDelta--
					return nil
				},
				Controller: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) uuid.UUID {
					return src.Controller
				},
				Label: "Vizier of Remedies: one fewer -1/-1 counter",
			},
		},
	})
}

// minusOneCounterPlacementOnYourCreature is Vizier of Remedies'
// predicate — plusOneCounterPlacementOnYourCreature's mirror on -1/-1
// counters, both placement-only (CR 122.6 / CR 614.1: a removal is not
// a placement) and both scoped to a creature the source's controller
// controls.
func minusOneCounterPlacementOnYourCreature(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
	if ev.Kind != game.RepEventCounter {
		return false
	}
	if ev.CounterDelta <= 0 {
		return false
	}
	if ev.CounterName != game.CounterMinusOne {
		return false
	}
	target, ok := g.LookupCardForEffect(ev.CounterTarget)
	if !ok {
		return false
	}
	if !target.IsCreature() {
		return false
	}
	return target.Controller == src.Controller
}
