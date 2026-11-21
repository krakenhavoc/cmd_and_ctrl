package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Test of Conviction — Sorcery {1}{W} (Star Trek, #4973):
//
//	"Choose one —
//	 • Battle Stations — Put a +1/+1 counter on each creature you control.
//	 • Begin Negotiations — Target opponent draws a card. You draw two
//	   cards."
//
// Battle Stations reads "each creature you control" as it resolves
// (CR 611.2c). Begin Negotiations is two draws in printed order, the
// second waiting for any prompt the first asks (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "beffaea5-beeb-4d4e-aca3-82c698caf6f8",
		Name:         "Test of Conviction",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.Mode("Battle Stations — Put a +1/+1 counter on each creature you control."),
			effects.ModeWithPurpose(effects.Mode("Begin Negotiations — Target opponent draws a card. You draw two cards.",
				effects.TargetPlayer("target opponent", effects.Opponent())), game.Purpose{Draws: 2}),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.HasMode(0) {
					return nil
				}
				return effects.PutAPlusOneCounterOnEachCreatureYouControlStep(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				item := ctx.Item
				if !ctx.HasMode(1) || len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer ||
					!ctx.IsTargetLegal(item.Targets[0]) {
					return nil
				}
				return effects.DrawCards{Player: item.Targets[0].ID, N: 1}.Apply(ctx)
			}),
			effects.YouDrawIfMode(1, 2),
		),
	})
}
