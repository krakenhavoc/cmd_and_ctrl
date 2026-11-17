package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Machine Over Matter — Instant for {1}{U} (The Brothers' War):
//
//	"This spell costs {1} less to cast if you control an artifact creature.
//	Return target nonland permanent to its owner's hand."
func init() {
	Register(Spec{
		OracleID:     "37069c66-c17a-40ac-80f4-14ef49d18008",
		Name:         "Machine Over Matter",
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target nonland permanent", Nonland()),
		SelfCostModifiers: []game.CostModifier{
			CostsLess(1, "This spell costs {1} less to cast if you control an artifact creature.",
				YouControlA(And(Artifact(), Creature()))),
		},
		OnResolve: bounceTheTarget,
	})
}
