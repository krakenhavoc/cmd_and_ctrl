package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Longstalk Brawl — Sorcery {G} (Bloomburrow, #4467):
//
//	"Gift a tapped Fish (You may promise an opponent a gift as you cast
//	 this spell. If you do, they create a tapped 1/1 blue Fish creature
//	 token before its other effects.)
//	 Choose target creature you control and target creature you don't
//	 control. Put a +1/+1 counter on the creature you control if the gift
//	 was promised. Then those creatures fight each other."
//
// Malamet Battle Glyph with the gift as the counter's condition. The
// gift is the shared keyword (gift.go), which makes the Fish before the
// spell's own text (CR 702.174j). The counter goes on the creature you
// control only if it is still a legal target (CR 608.2b); if either
// target is illegal, neither creature fights (CR 701.14b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "adc8976a-0871-4fd6-9c4c-822ba9549b4b",
		Name:         "Longstalk Brawl",
		Completeness: CompletenessFull,
		Gift:         GiftATappedFish(),
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature you don't control", OpponentControls()),
		),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				mine, ok := ctx.ClauseTarget(0)
				if !ok || mine.Kind != game.TargetCard || !ctx.GiftPromised() {
					return nil
				}
				return AddCounter{Target: mine.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}),
			Instruction(fightTargetClauses),
		),
	})
}
