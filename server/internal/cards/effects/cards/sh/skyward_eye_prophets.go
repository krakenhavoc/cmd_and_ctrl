package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skyward Eye Prophets — Creature — Human Wizard {3}{G}{W}{U}, 3/3 (Commander
// 2013, S82 OLD b385):
//
//	"Vigilance
//	 {T}: Reveal the top card of your library. If it's a land card, put it onto
//	 the battlefield. Otherwise, put it into your hand."
//
// Coiling Oracle's reveal as an activated ability: a land goes onto the
// battlefield, anything else to hand (a revealed token is neither put nor moved,
// CR 111.8; an empty library does nothing).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "45bef776-121b-4489-9c46-f7b4fd4c3c0d",
		Name:            "Skyward Eye Prophets",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Reveal the top card of your library. If it's a land card, put it onto the battlefield. Otherwise, put it into your hand.",
			Cost:    effects.TapCost(),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  effects.OldB322RevealTopLandElseHand,
		}},
	})
}
