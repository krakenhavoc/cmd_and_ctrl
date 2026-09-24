package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Three Steps Ahead — Instant {U}:
//
//	"Spree (Choose one or more additional costs.)
//	 + {1}{U} — Counter target spell.
//	 + {3} — Create a token that's a copy of target artifact or
//	   creature you control.
//	 + {2} — Draw two cards, then discard a card."
//
// The deck's proof card for Spree (CR 702.172a, ADR 0065's 2026-09-23
// amendment, issue #1330, tracker #888). Three bullets, three
// unrelated effects and three unrelated targets, so each is its own
// SpreeModeDoing closure rather than a branch in one OnResolve
// if-chain (Mishra's Command's pattern for "choose two" with per-mode
// targets).
//
// Casting all three bullets announces {U} (printed) + {1}{U} + {3} +
// {2} = {6}{U}{U} and three separate targets — a counterspell, a
// token-copy target and no target at all for the draw bullet — all
// legal to differ, and each re-checked against its OWN clause at
// resolution (ADR 0065 §2).
//
// The discard is the player's OWN choice (CR 701.8a): draw two, THEN
// queue the discard prompt, so both freshly drawn cards are legal to
// keep or discard, exactly as Stern Lesson's identical clause does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "282dfeaa-6243-4f92-838a-5cb54fa85184",
		Name:         "Three Steps Ahead",
		Completeness: CompletenessFull,
		Modes: Spree(
			SpreeModeDoing("Counter target spell.", "{1}{U}",
				TargetSpell("target spell"),
				CounterTheModesTarget),
			SpreeModeDoing("Create a token that's a copy of target artifact or creature you control.", "{3}",
				TargetPermanent("target artifact or creature you control", And(Or(Artifact(), Creature()), YouControl())),
				TokenCopyTheModesTarget),
			SpreeModeDoing("Draw two cards, then discard a card.", "{2}", nil,
				func(item *game.StackItem, ctx *Context, occ int) error {
					if err := (DrawCards{Player: item.Controller, N: 2}).Apply(ctx); err != nil {
						return err
					}
					ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
						Player: item.Controller,
						Source: item.SourceCardID,
						N:      1,
					})
					return nil
				}),
		),
	})
}
