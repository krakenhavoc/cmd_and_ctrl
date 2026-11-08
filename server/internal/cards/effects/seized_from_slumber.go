package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seized from Slumber — Instant {4}{W} (Duskmourn: House of Horror,
// #4466):
//
//	"This spell costs {3} less to cast if it targets a tapped creature.
//	 Destroy target creature."
//
// Ajani's Response, word for word: the reduction reads the announced
// target (CostsLessIfItTargets with hobB3TappedCreature, CR 601.2f), so
// whether the creature is tapped is judged as the cost is determined.
// The destroy is Bedevil's body (destroyTheTargetPermanent); a target
// that is no longer legal leaves the spell nothing to do (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7ac608bc-dc2a-43c9-9b3b-1b985c121d60",
		Name:         "Seized from Slumber",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature"),
		SelfCostModifiers: []game.CostModifier{
			CostsLessIfItTargets(3, "This spell costs {3} less to cast if it targets a tapped creature.", hobB3TappedCreature()),
		},
		OnResolve: destroyTheTargetPermanent,
	})
}
