package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "beffaea5-beeb-4d4e-aca3-82c698caf6f8",
		Name:         "Test of Conviction",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Battle Stations — Put a +1/+1 counter on each creature you control."),
			ModeWithPurpose(Mode("Begin Negotiations — Target opponent draws a card. You draw two cards.",
				TargetPlayer("target opponent", Opponent())), game.Purpose{Draws: 2}),
		),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				if !ctx.HasMode(0) {
					return nil
				}
				return putAPlusOneCounterOnEachCreatureYouControlStep(ctx)
			}),
			Instruction(func(ctx *Context) error {
				item := ctx.Item
				if !ctx.HasMode(1) || len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer ||
					!ctx.IsTargetLegal(item.Targets[0]) {
					return nil
				}
				return DrawCards{Player: item.Targets[0].ID, N: 1}.Apply(ctx)
			}),
			youDrawIfMode(1, 2),
		),
	})
}
