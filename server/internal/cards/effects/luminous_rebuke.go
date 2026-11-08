package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Luminous Rebuke — Instant {4}{W} (Foundations, #4465):
//
//	"This spell costs {3} less to cast if it targets a tapped creature.
//	 Destroy target creature."
//
// Seized from Slumber's text: CostsLessIfItTargets with the
// tapped-creature predicate reads the announced target (CR 601.2f, after
// 601.2c), and destroyTheTargetPermanent destroys whatever the target is
// on resolution, or nothing if it is no longer legal (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d96141af-313f-4ea9-a249-d9eb22cd2431",
		Name:         "Luminous Rebuke",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature"),
		SelfCostModifiers: []game.CostModifier{
			CostsLessIfItTargets(3, "This spell costs {3} less to cast if it targets a tapped creature.", hobB3TappedCreature()),
		},
		OnResolve: destroyTheTargetPermanent,
	})
}
