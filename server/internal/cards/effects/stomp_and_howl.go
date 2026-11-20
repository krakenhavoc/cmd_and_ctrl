package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stomp and Howl — Sorcery {2}{G} (Dissension, #5598):
//
//	"Destroy target artifact and target enchantment."
//
// Two clauses (ADR 0065 §1), destroyed together as one event
// (destroyEachLegalTarget); one that left in response is skipped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7cc809ed-5ce1-40d1-896d-014ca8e2b5a6",
		Name:         "Stomp and Howl",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: Clauses(
			TargetPermanent("target artifact", Artifact()),
			TargetPermanent("target enchantment", Enchantment()),
		),
		OnResolve: destroyEachLegalTarget,
	})
}
