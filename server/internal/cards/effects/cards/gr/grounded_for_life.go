package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grounded for Life — {4}{W} Instant (Teenage Mutant Ninja Turtles, #4457):
//
//	"This spell costs {3} less to cast if it targets a tapped creature.
//	 Destroy target creature."
//
// Ajani's Response's text under another name: CostsLessIfItTargets with
// the tapped-creature predicate reads the announced target (CR 601.2f),
// and the destroy is the shared destroyFirstLegalTarget, which does
// nothing if the target is no longer legal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9a334434-072a-4486-810a-63efb29dea6d",
		Name:         "Grounded for Life",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessIfItTargets(3, "This spell costs {3} less to cast if it targets a tapped creature.", effects.HobB3TappedCreature()),
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyFirstLegalTarget(ctx.Game, item)
		},
	})
}
