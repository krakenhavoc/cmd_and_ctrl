package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "b7aa0625-c691-4219-b099-3798b349d09c",
		Name:         "Stromgald Spy",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenAttacksAndIsNotBlockedEffect(label, func(g *game.Game, item *game.StackItem, defender uuid.UUID) error {
				return effects.MayChoice{
					Question: label + "? (it then assigns no combat damage this turn)",
					Then:     stromgaldSpyAnswer,
					Carry:    []uuid.UUID{defender},
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}

// stromgaldSpyAnswer is the "yes": the defending player (the carried ID)
// plays with their hand revealed for as long as the Spy stays, and the Spy
// assigns no combat damage this turn. Keyed (#4291).
var stromgaldSpyAnswer = effects.MayYes("confirm/stromgald-spy-reveal-hand", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	defender := effects.CarryAt(r, 0)
	if effects.DefendingPlayerStillIn(ctx.Game, defender) {
		if d, ok := effects.DurationWhileSourceRemains(ctx, ctx.Source()); ok {
			ctx.Game.GrantHandRevealedForEffect(defender, "Stromgald Spy", ctx.Source(), d)
		}
	}
	effects.AssignsNoCombatDamageThisTurn(ctx, "Stromgald Spy — assigns no combat damage this turn")
	return nil
})
