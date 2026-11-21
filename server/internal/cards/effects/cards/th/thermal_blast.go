package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thermal Blast — Instant {4}{R} (Odyssey, #5613):
//
//	"Thermal Blast deals 3 damage to target creature.
//	 Threshold — Thermal Blast deals 5 damage instead if there are seven
//	 or more cards in your graveyard."
//
// Threshold is an ability word (CR 207.2c); the count is read as the
// spell resolves, while Thermal Blast itself is still on the stack and
// so not one of the seven. The declared amount is the printed 3, the
// least it deals.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "52ad1a31-3ad8-4ead-ba77-675cad7e4a14",
		Name:         "Thermal Blast",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			amount := 3
			if effects.FraThresholdMet(ctx.Game, item.Controller) {
				amount = 5
			}
			return effects.DamageToFirstTargetStep(amount)(item, ctx)
		},
	})
}
