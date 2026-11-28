package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flare — Instant {2}{R} (Mirage):
//
//	"Flare deals 1 damage to any target.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// One damage to any target, then the delayed draw (CR 603.7) set up as the
// spell resolves; it still draws if the target has gone, so long as the spell
// resolved at all (CR 608.2b).
//
// No simplification.
func init() {
	purpose := effects.ForTargets(effects.DamageToTarget(0, 1))
	purpose.Draws = 1
	effects.Register(effects.Spec{
		OracleID:     "ff8fae94-3a34-4bc6-aaa2-7ab7d2e2d0dd",
		Name:         "Flare",
		Completeness: effects.CompletenessFull,
		Purpose:      purpose,
		Targets:      effects.TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.DamageToFirstTargetStep(1)(item, ctx); err != nil {
				return err
			}
			return effects.DrawACardAtTheNextUpkeep("Flare — draw a card").Apply(ctx)
		},
	})
}
