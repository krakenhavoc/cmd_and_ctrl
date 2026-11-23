package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stomp and Howl — Sorcery {2}{G} (Dissension, #5598):
//
//	"Destroy target artifact and target enchantment."
//
// Two clauses (ADR 0065 §1), destroyed together as one event
// (destroyEachLegalTarget); one that left in response is skipped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7cc809ed-5ce1-40d1-896d-014ca8e2b5a6",
		Name:         "Stomp and Howl",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.Clauses(
			effects.TargetPermanent("target artifact", effects.Artifact()),
			effects.TargetPermanent("target enchantment", effects.Enchantment()),
		),
		OnResolve: effects.DestroyEachLegalTarget,
	})
}
