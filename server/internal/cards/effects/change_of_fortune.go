package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Change of Fortune — Sorcery {3}{R}:
//
//	"Discard your hand, then draw a card for each card you've
//	 discarded this turn."
//
// The count is "this turn", not "just now" — a hand already thinned
// by a Faithless Looting earlier in the turn pays off here, and the
// cards THIS spell just discarded are themselves part of the count
// (the discard happens first, then the tally is read). #586's rule
// applies: "this turn" is read off Game.TurnTallyFor, never by
// walking g.Events. PlayerTurnTally had no discard counter yet, so
// this PR adds CardsDiscarded — one field on the tally plus one case
// in turnTallyListener (EventDiscardCard), the sanctioned way to
// widen it (AGENTS.md §7) rather than a card-side scan.
//
// "Discard your hand" takes every card, so there is no choice to
// make and the engine's random-selection discard (discardWholeHand,
// Ox of Agonas's helper) is exactly the printed instruction, not a
// simplification of it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "925feb7e-7876-430b-b95a-92b236a2e409",
		Name:         "Change of Fortune",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			caster := item.Controller
			if _, err := discardWholeHand(ctx.Game, caster); err != nil {
				return err
			}
			n := ctx.Game.TurnTallyFor(caster).CardsDiscarded
			return DrawCards{Player: caster, N: n}.Apply(ctx)
		},
	})
}
