package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "52ad1a31-3ad8-4ead-ba77-675cad7e4a14",
		Name:         "Thermal Blast",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			amount := 3
			if fraThresholdMet(ctx.Game, item.Controller) {
				amount = 5
			}
			return damageToFirstTarget(amount)(item, ctx)
		},
	})
}
