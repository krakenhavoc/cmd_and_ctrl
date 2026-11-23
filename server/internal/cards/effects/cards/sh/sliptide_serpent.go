package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sliptide Serpent — Creature — Serpent {4}{U}{U}, 4/4 (Nemesis, #5618):
//
//	"{3}{U}: Return this creature to its owner's hand."
//
// Aethertide Whale's self-bounce for mana: the shared
// returnThisPermanentToOwnersHand body behind {3}{U}. It goes to its
// OWNER's hand, so a stolen Serpent goes home. No tap in the cost, so it
// can be activated the turn it arrives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "25ae99c5-83b7-428d-b049-8ad40e250696",
		Name:         "Sliptide Serpent",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{U}: Return this creature to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{3}{U}"),
			Effect:  effects.ReturnThisPermanentToOwnersHand,
		}},
	})
}
