package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Etched Host Doombringer — Creature — Phyrexian Demon {4}{B}, 3/5 (March
// of the Machine):
//
//	"When this creature enters, choose one —
//	 • Target opponent loses 2 life and you gain 2 life.
//	 • Choose target battle. If an opponent protects it, remove three
//	   defense counters from it. Otherwise, put three defense counters
//	   on it."
//
// A modal enters trigger (Apothecary Stomper's shape): the mode and its
// target are chosen as the trigger goes on the stack. The protector is
// read as the second mode resolves.
//
// No simplification.
func init() {
	trigger := effects.WhenThisEnters("Etched Host Doombringer — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = effects.ChooseOne(
		effects.ModeDoing("Target opponent loses 2 life and you gain 2 life.",
			effects.TargetPlayer("target opponent", effects.Opponent()),
			func(item *game.StackItem, ctx *effects.Context, occ int) error {
				t, ok := effects.ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				if err := (effects.GainLife{Player: t.ID, Amount: -2}).Apply(ctx); err != nil {
					return err
				}
				return effects.GainLife{Player: item.Controller, Amount: 2}.Apply(ctx)
			}),
		effects.ModeDoing("Choose target battle. If an opponent protects it, remove three defense counters from it. Otherwise, put three defense counters on it.",
			effects.S82MomB12TargetBattle(),
			func(_ *game.StackItem, ctx *effects.Context, occ int) error {
				t, ok := effects.ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return effects.S82MomB12DefenseEither(ctx, t.ID, 3)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:     "85c799e9-74d5-459e-b168-b95ed78cd4e9",
		Name:         "Etched Host Doombringer",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
	})
}
