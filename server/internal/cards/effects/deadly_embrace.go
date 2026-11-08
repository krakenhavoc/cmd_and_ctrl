package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deadly Embrace — Sorcery {3}{B}{B} (Final Fantasy, #4462):
//
//	"Destroy target creature an opponent controls. Then draw a card for
//	 each creature that died this turn."
//
// Bake into a Pie's destroy, then a draw counted as that step runs (CR
// 608.2h) from the turn tally's table-wide death count (Mahadi, Emporium
// Master's b11CreaturesDiedThisTurn): any player's creature, a token
// included, and the one this spell just destroyed. A creature that
// survives the destroy (indestructible, regenerated) is not counted. The
// draw waits for any prompt the destroy asks (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "32ad99a9-b5ed-41a1-8060-22265f7c5da1",
		Name:         "Deadly Embrace",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature an opponent controls", OpponentControls()),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return destroyTheTargetPermanent(ctx.Item, ctx) }),
			Instruction(func(ctx *Context) error {
				return DrawCards{Player: ctx.Controller(), N: b11CreaturesDiedThisTurn(ctx.Game)}.Apply(ctx)
			}),
		),
	})
}
