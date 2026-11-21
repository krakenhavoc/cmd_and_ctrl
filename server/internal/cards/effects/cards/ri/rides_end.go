package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ride's End — Instant {4}{W} (Aetherdrift, #4464):
//
//	"This spell costs {3} less to cast if it targets a tapped permanent.
//	 Exile target creature or Vehicle."
//
// Ajani's Response's reduction (CostsLessIfItTargets) over a tapped
// permanent rather than a tapped creature, judged against the announced
// target when the cost is determined (CR 601.2f, after 601.2c), so an
// uncrewed Vehicle that is tapped counts. The exile is
// exileFirstLegalCardTarget: a target that is no longer a legal one on
// resolution is left alone (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "89e209f3-c474-453f-9659-ccebfc88bc0f",
		Name:         "Ride's End",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target creature or Vehicle", effects.Or(effects.Creature(), effects.HasSubtype("Vehicle"))),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessIfItTargets(3, "This spell costs {3} less to cast if it targets a tapped permanent.", effects.RfCreatureDTapped()),
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ExileFirstLegalCardTarget(ctx.Game, item)
		},
	})
}
