package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "1b73639b-2b07-4e82-b990-ca7721074e43",
		Name:         "Dissenter's Deliverance",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact", effects.Artifact()),
		OnResolve:    effects.DestroyTheTargetPermanent,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{G}")},
	})
}
