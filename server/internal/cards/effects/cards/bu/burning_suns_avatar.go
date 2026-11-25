package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Burning Sun's Avatar — Creature — Dinosaur Avatar {3}{R}{R}{R}, 6/6
// (Ixalan, #5532):
//
//	"When this creature enters, it deals 3 damage to target opponent or
//	 planeswalker and 3 damage to up to one target creature."
//
// Two clauses chosen as the trigger goes on the stack: the opponent or
// planeswalker, and up to one creature (any controller, the Avatar
// included). Each is dealt 3 only if it is still legal as the trigger
// resolves (CR 608.2b), together as one damage instruction.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "248d7759-aa73-4430-aa86-d45dc607944b",
		Name:         "Burning Sun's Avatar",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Burning Sun's Avatar — 3 damage to target opponent or planeswalker and 3 damage to up to one target creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					var targets []uuid.UUID
					for slot := 0; slot < 2; slot++ {
						if t, ok := ctx.ClauseTarget(slot); ok {
							targets = append(targets, t.ID)
						}
					}
					return effects.DealDamageToEachThen(ctx, targets, 3, effects.RecipientThen{})
				}),
				effects.Clauses(effects.TargetOpponentOrPlaneswalker(),
					effects.TargetCreature("up to one target creature").WithCount(0, 1))),
				effects.ForTargets(effects.DamageToTarget(0, 3), effects.DamageToTarget(1, 3))),
		},
	})
}
