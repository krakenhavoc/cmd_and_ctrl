package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Assassinate — Sorcery {2}{B} (Time Spiral, #5596):
//
//	"Destroy target tapped creature."
//
// Vengeance's body at sorcery speed for three: the target must be tapped
// at announce and still be tapped at resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ce979f36-84f2-4419-b400-971263494cc6",
		Name:         "Assassinate",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target tapped creature", effects.TappedPermanent()),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
