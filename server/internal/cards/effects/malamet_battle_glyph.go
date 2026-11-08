package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Malamet Battle Glyph — Sorcery {G} (The Lost Caverns of Ixalan, #4470):
//
//	"Choose target creature you control and target creature you don't
//	 control. If the creature you control entered this turn, put a +1/+1
//	 counter on it. Then those creatures fight each other."
//
// Hog-Monkey Rampage's two clauses with a different condition on the
// counter: the turn tally's record of what entered this turn
// (Game.EnteredThisTurn), read as the spell resolves. A creature you
// control that is no longer a legal target gets no counter (CR 608.2b);
// if either target is illegal, neither creature fights (CR 701.14b), but
// a legal creature you control still gets its counter. "You don't
// control" is OpponentControls, as every permanent you don't control is
// an opponent's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "483a0caa-e912-41e9-aaa3-bcea9369ec06",
		Name:         "Malamet Battle Glyph",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature you don't control", OpponentControls()),
		),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				mine, ok := ctx.ClauseTarget(0)
				if !ok || mine.Kind != game.TargetCard || !ctx.Game.EnteredThisTurn(mine.ID) {
					return nil
				}
				return AddCounter{Target: mine.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}),
			Instruction(fightTargetClauses),
		),
	})
}
