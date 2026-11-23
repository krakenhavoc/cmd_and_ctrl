package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Galvanize — Instant {1}{R} (Murders at Karlov Manor, #4469):
//
//	"Galvanize deals 3 damage to target creature. If you've drawn two or
//	 more cards this turn, Galvanize deals 5 damage to that creature
//	 instead."
//
// The draw count is the caster's turn tally (CondCardsDrawn, Lake-town
// Toymaker's count), read as the spell resolves, so a card drawn in
// response still counts. A target that left in response takes nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9b44c005-6799-47fd-8b0d-7179b5281d78",
		Name:         "Galvanize",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			target := effects.FirstLegalBattlefieldTarget(ctx)
			if target == uuid.Nil {
				return nil
			}
			amount := 3
			if (effects.CondCardsDrawn{AtLeast: 2}).Holds(ctx) {
				amount = 5
			}
			return effects.DealDamage{Source: ctx.Source(), Target: target, Amount: amount}.Apply(ctx)
		},
	})
}
