package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Retreat to Kazandu — {2}{G} Enchantment: "Landfall — Whenever a land
// you control enters, choose one — • Put a +1/+1 counter on target
// creature. • You gain 2 life."
func init() {
	t := effects.Landfall("Retreat to Kazandu — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = effects.ChooseOne(
		effects.ModeDoing("Put a +1/+1 counter on target creature.", effects.TargetCreature("target creature"),
			func(_ *game.StackItem, ctx *effects.Context, occ int) error {
				tgt, ok := effects.ModeTarget(ctx, occ)
				if !ok || tgt.Kind != game.TargetCard {
					return nil
				}
				return effects.AddCounter{Target: tgt.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}),
		effects.ModeDoing("You gain 2 life.", nil,
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.GainLife{Amount: 2}.Apply(ctx)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:     "3f8e5ff1-af89-427e-924c-19a44f9a3788",
		Name:         "Retreat to Kazandu",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
