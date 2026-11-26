package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Faerie Miscreant — Creature — Faerie Rogue {U}, 1/1 (Core Set 2020, #5518):
//
//	"Flying
//	 When this creature enters, if you control another creature named Faerie
//	 Miscreant, draw a card."
//
// An intervening "if" (CR 603.4): the trigger is raised only while another
// creature of that name is under your control and does nothing if none is as
// it resolves. The Miscreant itself never counts as "another".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dd2510cd-356c-4d5a-8e5d-e44872eba60b",
		Name:            "Faerie Miscreant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.OldB154IfYouControlAnotherNamed("Faerie Miscreant",
				effects.WhenThisEnters("Faerie Miscreant — draw a card", effects.Do(effects.DrawCards{N: 1}))),
		},
	})
}
