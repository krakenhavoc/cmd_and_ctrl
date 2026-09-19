package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Syphon Mind — Sorcery {3}{B} (EDHREC rank 1049):
//
//	"Each other player discards a card. You draw a card for each card
//	 discarded this way."
//
// The multiplayer Sign in Blood: three cards for four mana at a full
// table. Each opponent's discard is their own choice, one prompt per
// opponent over their own hand, so nobody picks for anyone else.
//
// The catalog's one card that reads discard-FIRST-then-count, and
// since #1027 it is written as the engine's prompted-discard RUN:
// "each other player discards a card" is ONE instruction however many
// prompts it takes, and "a card for each card discarded THIS WAY" is
// its continuation, run once after the last opponent has answered and
// their cards have finished moving.
//
// Three things that buys over the per-prompt Then it used to ride:
//
//   - The draw happens ONCE, for the whole table, after the table has
//     discarded — which is the printed card. It used to arrive one
//     card per answer, and that was a declared cosmetic simplification.
//   - It counts what CR 701.8a counts. A madness card exiled instead
//     of binned (CR 702.35a) was still discarded and still buys a
//     card; a leg the CR 614 window cancelled did not and does not.
//     The old count was "how many prompts had a Then", which is
//     neither.
//   - The empty-hand skip is the engine's rule, not this card's. An
//     opponent with nothing in hand is never asked (CR 701.8a discards
//     as many as you can), gets no entry in the run's answer, and buys
//     the caster nothing — which used to need a guard here, written
//     precisely because the prompt's own Then fires for an empty hand
//     and this card's "then" is a payoff rather than the next
//     instruction.
//
// An opponent who leaves the game before answering buys nothing: the
// departure settles their leg with nothing discarded (#1016's
// dropDefault, ADR 0013 §5y).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "abc37d6c-6300-47b5-a679-9db5b83eb54f",
		Name:         "Syphon Mind",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			caster := item.Controller
			return ctx.Game.EachPlayerDiscardsThenForEffect(caster,
				game.DiscardPrompt{Source: item.SourceCardID, N: 1},
				func(g *game.Game, discarded game.PromptedDiscards) error {
					if n := discarded.Count(); n > 0 {
						return g.DrawNForEffect(caster, n)
					}
					return nil
				})
		},
	})
}
