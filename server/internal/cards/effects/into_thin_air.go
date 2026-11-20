package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Into Thin Air — Instant {5}{U} (Fifth Dawn, #5604):
//
//	"Affinity for artifacts (This spell costs {1} less to cast for each
//	 artifact you control.)
//	 Return target artifact to its owner's hand."
//
// Affinity for artifacts (the shared self cost modifier) on Boomerang's
// bounce, limited to an artifact.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "06df923f-916e-49f4-960a-7b3fb108c704",
		Name:         "Into Thin Air",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for artifacts", Artifact()),
		},
		Targets:   TargetPermanent("target artifact", Artifact()),
		Purpose:   game.Purpose{Answers: game.AnswerRemove},
		OnResolve: bounceTheTarget,
	})
}
