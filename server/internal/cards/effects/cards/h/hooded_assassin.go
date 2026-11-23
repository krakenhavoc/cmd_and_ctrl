package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hooded Assassin — Creature — Human Assassin {2}{B}, 1/2 (Jumpstart
// 2022, #5948):
//
//	"When this creature enters, choose one —
//	 • Put a +1/+1 counter on this creature.
//	 • Destroy target creature that was dealt damage this turn."
//
// Qutrub Forayer's modal enters trigger and "dealt damage this turn"
// target filter, with Blade of the Swarm's counter bullet. The mode is
// chosen as the trigger goes on the stack (CR 700.2b), so the destroy
// bullet is only chosen with a damaged creature to aim at.
//
// No simplification.
func init() {
	trigger := effects.WhenThisEnters("Hooded Assassin — choose one", func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = effects.ChooseOne(
		effects.ModeDoing("Put a +1/+1 counter on this creature.", nil,
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}),
		effects.ModeDoing("Destroy target creature that was dealt damage this turn.",
			effects.TargetCreature("target creature that was dealt damage this turn", effects.DealtDamageThisTurn()),
			effects.DestroyTheModesTarget),
	)
	effects.Register(effects.Spec{
		OracleID:     "a3f76c99-b608-425e-83ca-675acb7c2693",
		Name:         "Hooded Assassin",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
	})
}
