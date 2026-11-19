package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Retreat to Kazandu — {2}{G} Enchantment: "Landfall — Whenever a land
// you control enters, choose one — • Put a +1/+1 counter on target
// creature. • You gain 2 life."
func init() {
	t := Landfall("Retreat to Kazandu — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = ChooseOne(
		ModeDoing("Put a +1/+1 counter on target creature.", TargetCreature("target creature"),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				tgt, ok := ModeTarget(ctx, occ)
				if !ok || tgt.Kind != game.TargetCard {
					return nil
				}
				return AddCounter{Target: tgt.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}),
		ModeDoing("You gain 2 life.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return GainLife{Amount: 2}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:     "3f8e5ff1-af89-427e-924c-19a44f9a3788",
		Name:         "Retreat to Kazandu",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
