package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stromgald Spy — Creature — Human Rogue {3}{B}, 2/4:
//
//	"Whenever this creature attacks and isn't blocked, you may have
//	 defending player play with their hand revealed for as long as this
//	 creature remains on the battlefield. If you do, this creature assigns
//	 no combat damage this turn."
//
// #3953, ADR 0199: the defending player is the one whose block
// declaration fired the trigger. Saying yes grants them a revealed hand
// (game.GrantHandRevealedForEffect) for as long as this Spy, as the object
// it is now, stays on the battlefield (CR 611.2b): a Spy that left before
// the trigger resolved reveals nothing. Every seat then sees the cards in
// that hand, including the ones drawn later, until the Spy leaves.
//
// No simplifications.
func init() {
	const label = "Stromgald Spy — you may have defending player play with their hand revealed"
	Register(Spec{
		OracleID:     "b7aa0625-c691-4219-b099-3798b349d09c",
		Name:         "Stromgald Spy",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenAttacksAndIsNotBlockedEffect(label, func(g *game.Game, item *game.StackItem, defender uuid.UUID) error {
				return MayChoice{
					Question: label + "? (it then assigns no combat damage this turn)",
					OnYes: func(ctx *Context) error {
						if defendingPlayerStillIn(ctx.Game, defender) {
							if d, ok := DurationWhileSourceRemains(ctx, ctx.Source()); ok {
								ctx.Game.GrantHandRevealedForEffect(defender, "Stromgald Spy", ctx.Source(), d)
							}
						}
						assignsNoCombatDamageThisTurn(ctx, "Stromgald Spy — assigns no combat damage this turn")
						return nil
					},
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
