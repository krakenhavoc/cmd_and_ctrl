package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Galadriel, Gift-Giver — Legendary Creature — Elf Noble {3}{G}{G}, 4/4 (The
// Lord of the Rings: Tales of Middle-earth):
//
//	"Whenever Galadriel enters or attacks, choose one —
//	 • Put a +1/+1 counter on another target creature.
//	 • Create a Food token.
//	 • Create a Treasure token."
//
// One modal trigger watching two events (WhenThisEntersOrAttacks), the mode
// and its target chosen as it goes on the stack (CR 603.3c). The first mode's
// target is "another" creature (never Galadriel); a mode with no legal
// target can't be chosen (CR 603.3c), which leaves the two token modes.
//
// No simplification.
func init() {
	trig := effects.WhenThisEntersOrAttacks("Galadriel, Gift-Giver — choose one",
		func(_ *game.Game, _ *game.StackItem) error { return nil })
	trig.Modes = effects.ChooseOne(
		effects.ModeDoing("Put a +1/+1 counter on another target creature.", effects.Another(effects.TargetCreature("another target creature")),
			func(_ *game.StackItem, ctx *effects.Context, occ int) error {
				t, ok := effects.ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}),
		effects.ModeDoing("Create a Food token.", nil, func(_ *game.StackItem, ctx *effects.Context, _ int) error {
			return effects.CreateToken{Template: effects.FoodToken(), N: 1}.Apply(ctx)
		}),
		effects.ModeDoing("Create a Treasure token.", nil, func(_ *game.StackItem, ctx *effects.Context, _ int) error {
			return effects.CreateToken{Template: effects.TreasureToken(), N: 1}.Apply(ctx)
		}),
	)
	effects.Register(effects.Spec{
		OracleID:     "d0c0a2ef-27b4-4186-8538-6ba5d35d85bb",
		Name:         "Galadriel, Gift-Giver",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{trig},
	})
}
