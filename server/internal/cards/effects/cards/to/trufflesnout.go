package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trufflesnout — Creature — Boar {2}{G}, 2/2:
//
//	"When this creature enters, choose one —
//	 • Put a +1/+1 counter on this creature.
//	 • You gain 4 life."
//
// A modal enters trigger (CR 700.2): the bullet is chosen as the
// trigger is put on the stack. If Trufflesnout has left by the time the
// counter bullet resolves, nothing gets the counter (CR 400.7).
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters("Trufflesnout — choose one", func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = effects.ChooseOne(
		effects.ModeDoing("Put a +1/+1 counter on this creature.", nil, effects.PutAPlusOneCounterOnThisCreatureMode),
		effects.ModeDoing("You gain 4 life.", nil, func(item *game.StackItem, ctx *effects.Context, _ int) error {
			return effects.GainLife{Player: item.Controller, Amount: 4}.Apply(ctx)
		}),
	)
	effects.Register(effects.Spec{
		OracleID:     "8a15841a-dcb6-454a-b865-43a067f42b4e",
		Name:         "Trufflesnout",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
	})
}
