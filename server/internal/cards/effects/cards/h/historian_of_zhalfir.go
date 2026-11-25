package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Historian of Zhalfir — Creature — Human Wizard {2}{U}{U}, 3/3 (Core Set
// 2021, #5511):
//
//	"Whenever this creature attacks, if you control a Teferi planeswalker,
//	 draw a card."
//
// An intervening "if" (CR 603.4): checked as the creature attacks and again
// as the trigger resolves. "A Teferi planeswalker" is the planeswalker with
// the Teferi subtype, not a card merely named Teferi (Liliana's Triumph's
// reading).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "87e9065d-814a-47fa-9e92-19fb6fd6d78c",
		Name:         "Historian of Zhalfir",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("you control a Teferi planeswalker",
				func(g *game.Game, you uuid.UUID) bool { return effects.B42ControlsPlaneswalkerNamed(g, you, "Teferi") },
				effects.WheneverThisAttacks("Historian of Zhalfir — draw a card", effects.Do(effects.DrawCards{N: 1}))),
		},
	})
}
