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
// "ON A PERMANENT OR PLAYER" is both halves of the card, and since
// ADR 0056 both halves are reachable: counters on players go through
// the same CR 614 window (game.ReplacementEvent.CounterPlayer), so
// Vorinclex doubles your proliferated experience counters and halves
// the poison an opponent's infect creature gives you.
//
// WHO IS PUTTING THEM is read off the event
// (game.ReplacementEvent.CounterPlacer), which is what the card
// actually says, and since #3942 every placement names one at its
// source (game/counter_placer.go): the controller of the resolving
// spell or ability, the player paying a cost (a loyalty ability's +N,
// which this card does apply to — it names no "effect"), the activator
// of a mana ability, the controller of a permanent entering with
// counters (CR 122.6a), the player whose Saga gets its lore counter
// (CR 714.3c), the seat making a sandbox edit. So your spell putting a
// -1/-1 counter on an opponent's creature is doubled, and an opponent's
// +1 on their own planeswalker or the lore counter on their Saga is
// halved to nothing. This shipped keyed on the counter TARGET's
// controller, which disagrees exactly when it matters most.
//
// Rounding is DOWN and floor division on a negative delta rounds the
// wrong way in Go, so the halving arm is guarded to placements only.
// A counter REMOVAL is not a placement and no replacement applies to
// it (CR 614.1) — which is also what keeps damage-driven loyalty loss
// out of this, since that path bypasses the pipeline entirely.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5a3fdf5a-bff8-4896-b288-3f43f9a72d9b",
		Name:            "Vorinclex, Monstrous Raider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Replacements: []game.ReplacementEffect{
			{
				Watches: []game.EventKind{game.EventCounterPlaced},
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
					return counterPlacementBy(ev, g, src.Controller, true)
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
					return counterPlacementBy(ev, g, src.Controller, false)
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

// counterPlacementBy reports whether `ev` is a counter PLACEMENT — on
// a permanent or on a player — whose PLACER is (or is not) `you`.
//
// It is the "IF YOU WOULD PUT" question, and it is NOT the same as
// counterPlacementOn below, which asks who the counters are going on.
// The two agree for the common case — a player putting counters on
// their own permanents — and disagree exactly where this card is
// interesting: an opponent's effect putting counters on your creature
// is a placement BY an opponent ON a permanent you control, and
// Vorinclex halves it.
//
// They were ONE helper until ADR 0056 gave the event a placer, which
// is how Pir, Imaginative Rascal — a passive "would be put" card with
// no "you" in its text — ended up sharing a predicate with a card that
// asks the opposite question.
//
// Placement only: a negative delta is a removal, and CR 614.1's
// replacement effects apply to events that WOULD HAPPEN, of which
// "remove three loyalty counters" is not one this card names. Halving
// a removal would also round the wrong way in Go, where integer
// division truncates toward zero.
func counterPlacementBy(ev *game.ReplacementEvent, _ *game.Game, you uuid.UUID, mine bool) bool {
	if ev.Kind != game.RepEventCounter || ev.CounterDelta <= 0 {
		return false
	}
	// The engine names the placer of every placement (#3942); Nil is
	// a placement that is going nowhere, and nobody is putting it.
	placer := ev.CounterPlacer
	if placer == uuid.Nil {
		return false
	}
	if mine {
		return placer == you
	}
	return placer != you
}

// counterPlacementOn reports whether `ev` is a counter PLACEMENT on a
// PERMANENT whose controller is `you`.
//
// The passive "if one or more counters would be put on <something you
// control>" reading — Pir, Imaginative Rascal — which names no placer
// and so asks nothing about one. It is deliberately blind to counters
// on PLAYERS: a permanent is not a player, and one event kind now
// carries both (ADR 0056 Decision 5), so the card that says "permanent"
// has to say so here. counter_placer_test.go holds every registered
// counter replacement to that.
func counterPlacementOn(ev *game.ReplacementEvent, g *game.Game, you uuid.UUID) bool {
	if ev.Kind != game.RepEventCounter || ev.CounterDelta <= 0 {
		return false
	}
	if ev.CounterPlayer != uuid.Nil {
		return false
	}
	target, ok := g.LookupCardForEffect(ev.CounterTarget)
	if !ok {
		return false
	}
	return target.Controller == you
}

func vorinclexController(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
	return src.Controller
}
