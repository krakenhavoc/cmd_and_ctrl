package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Broken Wings — Instant {2}{G} (Foundations, #4465):
//
//	"Destroy target artifact, enchantment, or creature with flying."
//
// One target clause, three kinds; "with flying" qualifies only the
// creature, as Crushing Canopy's creature bullet reads it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5e316864-d55c-496f-8f46-773567896864",
		Name:         "Broken Wings",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: TargetPermanent("target artifact, enchantment, or creature with flying",
			Or(Artifact(), Enchantment(), And(Creature(), HasKeyword("flying")))),
		OnResolve: destroyTheTargetPermanent,
	})
}
