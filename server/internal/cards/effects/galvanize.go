package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "9b44c005-6799-47fd-8b0d-7179b5281d78",
		Name:         "Galvanize",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(DamageToTarget(0, 3)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			target := FirstLegalBattlefieldTarget(ctx)
			if target == uuid.Nil {
				return nil
			}
			amount := 3
			if (CondCardsDrawn{AtLeast: 2}).Holds(ctx) {
				amount = 5
			}
			return DealDamage{Source: ctx.Source(), Target: target, Amount: amount}.Apply(ctx)
		},
	})
}
