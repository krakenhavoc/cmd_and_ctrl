package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Touch of Death — Sorcery {2}{B} (Ice Age, S82 OLD, #5636):
//
//	"Touch of Death deals 1 damage to target player or planeswalker. You gain 1
//	 life.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// You gain the life even if the target is gone only when the spell resolves at
// all (CR 608.2b): a spell with no legal target does nothing.
//
// No simplification.
func init() {
	purpose := effects.ForTargets(effects.DamageToTarget(0, 1))
	purpose.Draws = 1
	effects.Register(effects.Spec{
		OracleID:     "52a768eb-16eb-4643-b8be-7a4b5b32f28b",
		Name:         "Touch of Death",
		Completeness: effects.CompletenessFull,
		Purpose:      purpose,
		Targets:      effects.TargetPlayerOrPlaneswalker(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.DamageToFirstTargetStep(1)(item, ctx); err != nil {
				return err
			}
			if err := (effects.GainLife{Amount: 1}).Apply(ctx); err != nil {
				return err
			}
			return effects.DrawACardAtTheNextUpkeep("Touch of Death — draw a card").Apply(ctx)
		},
	})
}
