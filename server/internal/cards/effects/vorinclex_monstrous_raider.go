package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vorinclex, Monstrous Raider — 6/6 Phyrexian Praetor for {4}{G}{G}:
//
//	"Trample, haste
//	 If you would put one or more counters on a permanent or player,
//	 put twice that many of each of those kinds of counters on that
//	 permanent or player instead.
//	 If an opponent would put one or more counters on a permanent or
//	 player, they put half that many of each of those kinds of
//	 counters on that permanent or player instead, rounded down."
//
// Two CR 614 replacement effects on one card, pointing opposite ways,
// and the asymmetry is the card. Both read the COUNTER-PLACER, not
// the counter's target: Vorinclex doubles the loyalty you stamp onto
// an opponent's planeswalker with your own effect, and halves what an
// opponent puts on YOUR creature.
//
// SANDBOX GAP, and it is in the halving half: ReplacementEvent
// carries no "who is placing these counters" field, so the opponent
// clause is keyed on the counter TARGET's controller instead. The two
// agree for the overwhelmingly common case — a player putting
// counters on their own permanents — and disagree when one player's
// effect puts counters on another's permanent, where this halves what
// should have been left alone (or left alone what should have been
// halved).
//
// Weaker or unchanged, never stronger, in the case that matters: the
// doubling half is exact, because "you would put" and "a permanent
// you control" coincide for every card that combos with it — the
// proliferate, the +1/+1 engines, the loyalty stamp.
//
// Rounding is DOWN and floor division on a negative delta rounds the
// wrong way in Go, so the halving arm is guarded to placements only.
// A counter REMOVAL is not a placement and no replacement applies to
// it (CR 614.1) — which is also what keeps damage-driven loyalty loss
// out of this, since that path bypasses the pipeline entirely.
func init() {
	Register(Spec{
		OracleID:        "5a3fdf5a-bff8-4896-b288-3f43f9a72d9b",
		Name:            "Vorinclex, Monstrous Raider",
		PrintedKeywords: []string{"trample", "haste"},
		Replacements: []game.ReplacementEffect{
			{
				Watches: []game.EventKind{game.EventCounterPlaced},
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
					return counterPlacementOn(ev, g, src.Controller, true)
				},
				Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
					ev.CounterDelta *= 2
					return nil
				},
				Controller: vorinclexController,
				Label:      "Vorinclex: double your counters",
			},
			{
				Watches: []game.EventKind{game.EventCounterPlaced},
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
					return counterPlacementOn(ev, g, src.Controller, false)
				},
				Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
					ev.CounterDelta /= 2
					return nil
				},
				Controller: vorinclexController,
				Label:      "Vorinclex: halve opponents' counters",
			},
		},
	})
}

// counterPlacementOn reports whether `ev` is a counter PLACEMENT on a
// permanent whose controller is (or is not) `you`.
//
// Placement only: a negative delta is a removal, and CR 614.1's
// replacement effects apply to events that WOULD HAPPEN, of which
// "remove three loyalty counters" is not one this card names. Halving
// a removal would also round the wrong way in Go, where integer
// division truncates toward zero.
func counterPlacementOn(ev *game.ReplacementEvent, g *game.Game, you uuid.UUID, mine bool) bool {
	if ev.Kind != game.RepEventCounter || ev.CounterDelta <= 0 {
		return false
	}
	target, ok := g.LookupCardForEffect(ev.CounterTarget)
	if !ok {
		return false
	}
	if mine {
		return target.Controller == you
	}
	return target.Controller != you
}

func vorinclexController(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
	return src.Controller
}
