package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Take Vengeance — Sorcery {1}{W} (Core Set 2019, #5527):
//
//	"Destroy target tapped creature."
//
// Assassinate's text: destroy one tapped creature, checked again on
// resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c60b8edd-aa3f-48ea-b202-d10599f91d9d",
		Name:         "Take Vengeance",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target tapped creature", effects.TappedPermanent()),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
