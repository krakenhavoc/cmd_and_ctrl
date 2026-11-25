package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Drown in Shapelessness (Global Series Jiang Yanggu & Mu Yanling, #5528):
//
//	Return target creature to its owner's hand.
//
// {1}{U} Instant.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a3874282-9887-4b53-b80e-564ac99cca23",
		Name:         "Drown in Shapelessness",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve:    effects.BounceTheTarget,
	})
}
