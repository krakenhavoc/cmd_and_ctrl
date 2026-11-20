package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dissenter's Deliverance — Instant {1}{G}
// (Amonkhet, #5535):
//
//	"Destroy target artifact.
//	 Cycling {G} ({G}, Discard this card: Draw a card.)"
//
// Shatter, plus cycling.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1b73639b-2b07-4e82-b990-ca7721074e43",
		Name:         "Dissenter's Deliverance",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target artifact", Artifact()),
		OnResolve:    destroyTheTargetPermanent,
		Activated:    []ActivatedAbility{Cycling("{G}")},
	})
}
