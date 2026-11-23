package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Asphyxiate — Sorcery {1}{B}{B} (Born of the Gods,
// #5559):
//
//	"Destroy target untapped creature."
//
// Assassinate's body with the opposite tap clause: the predicate is on
// the target, so a tapped creature is never offered (CR 115.4) and one
// tapped in response makes the spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1c2ce0d0-8c3e-48e8-b1fd-54658bdb9af9",
		Name:         "Asphyxiate",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target untapped creature", effects.UntappedCreature()),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
