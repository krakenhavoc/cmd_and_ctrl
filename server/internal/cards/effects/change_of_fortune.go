package effects

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
// widen it (docs/adding-cards.md) rather than a card-side scan.
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
		// The draw waits for a prompt a discard asks (a replacement offer such as Library of Leng's; CR 608.2c, #4155) and
		// reads the turn's tally, not a count carried from the discard.
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				_, err := discardWholeHand(ctx.Game, ctx.Controller())
				return err
			}),
			Instruction(func(ctx *Context) error {
				caster := ctx.Controller()
				return DrawCards{Player: caster, N: ctx.Game.TurnTallyFor(caster).CardsDiscarded}.Apply(ctx)
			}),
		),
	})
}
