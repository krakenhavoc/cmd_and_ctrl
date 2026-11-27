package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Deadshot — Sorcery {3}{R} (Tempest, S82 OLD batch 757, #5628):
//
//	"Tap target creature. It deals damage equal to its power to another
//	 target creature."
//
// Two clauses read by slot (Bionic Blow's shape): the tap happens, then the
// first creature, tapped, bites the second with the power it has as the
// bite happens (CR 608.2h if it left). If the first creature is gone the
// bite deals nothing; the second still being legal is judged on its own
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "711c10cb-7a85-417d-be24-79d4e4a89993",
		Name:         "Deadshot",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // a one-sided bite after the tap
		Targets: effects.Clauses(
			effects.TargetCreature("target creature"),
			effects.Distinct(effects.TargetCreature("another target creature")),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(effects.S82OldB757TapFirstClauseTarget),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.OneSidedBite(ctx.Item, ctx)
			}),
		),
	})
}
