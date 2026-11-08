package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "89e209f3-c474-453f-9659-ccebfc88bc0f",
		Name:         "Ride's End",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target creature or Vehicle", Or(Creature(), HasSubtype("Vehicle"))),
		SelfCostModifiers: []game.CostModifier{
			CostsLessIfItTargets(3, "This spell costs {3} less to cast if it targets a tapped permanent.", rfCreatureDTapped()),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return exileFirstLegalCardTarget(ctx.Game, item)
		},
	})
}
