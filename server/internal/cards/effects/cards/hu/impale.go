package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Impale — Sorcery {2}{B}{B} (Rivals of Ixalan, #5531):
//
//	"Destroy target creature."
//
// Murder at sorcery speed: the body is the shared
// destroyTheTargetPermanent, and a target that left in response makes
// the spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dcdf40a9-5a7e-45b1-aee4-36dda598baaf",
		Name:         "Impale",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
