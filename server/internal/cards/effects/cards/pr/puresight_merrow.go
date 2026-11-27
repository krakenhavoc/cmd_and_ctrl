package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Puresight Merrow — Creature — Merfolk Wizard {W/U}{W/U}, 2/2 (Shadowmoor, S82
// OLD b531, #5591):
//
//	"{W/U}, {Q}: Look at the top card of your library. You may exile that card.
//	 ({Q} is the untap symbol.)"
//
// {Q} is AbilityCost.UntapSelf (CR 107.6). A look (CR 701.20a: only you learn
// the card), then a "you may": declining leaves it on top, accepting exiles it
// face up, so the table sees it go. An empty library does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "012c665c-24b1-49d6-b575-2b4ee1c3a833",
		Name:         "Puresight Merrow",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W/U}, {Q}: Look at the top card of your library. You may exile that card.",
			Cost:    effects.OldB531UntapCost("{W/U}"),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  effects.OldB531LookAtTopMayExile("Puresight Merrow"),
		}},
	})
}
