package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "32ad99a9-b5ed-41a1-8060-22265f7c5da1",
		Name:         "Deadly Embrace",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.DestroyTheTargetPermanent(ctx.Item, ctx) }),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DrawCards{Player: ctx.Controller(), N: effects.B11CreaturesDiedThisTurn(ctx.Game)}.Apply(ctx)
			}),
		),
	})
}
