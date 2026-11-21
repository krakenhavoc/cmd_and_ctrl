package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Violent Impact — Sorcery {3}{R}
// (Amonkhet, #5535):
//
//	"Destroy target artifact or land.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Sunder from Within, plus cycling.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d48d9e93-72a2-48f2-8164-8032615ee217",
		Name:         "Violent Impact",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target artifact or land", effects.Or(effects.Artifact(), effects.Land())),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve:    effects.DestroyTheTargetPermanent,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
