package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fate of the Sun-Cryst — Instant {4}{W} (Final Fantasy, #4462):
//
//	"This spell costs {2} less to cast if it targets a tapped creature.
//	 Destroy target nonland permanent."
//
// Void Rend's destroy behind Magnificent End's reduction: the cost reads
// the announced target (CR 601.2f, after 601.2c), so the discount
// applies when the nonland permanent chosen is a tapped creature, and
// not for a tapped artifact or an untapped creature. Whether it is
// tapped is judged when the cost is determined; the destroy reads the
// target on resolution (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0ffeec93-89d7-42c2-8abf-802e9f6dc992",
		Name:         "Fate of the Sun-Cryst",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target nonland permanent", Nonland()),
		SelfCostModifiers: []game.CostModifier{
			CostsLessIfItTargets(2, "This spell costs {2} less to cast if it targets a tapped creature.", hobB3TappedCreature()),
		},
		OnResolve: destroyTheTargetPermanent,
	})
}
