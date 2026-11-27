package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wormfang Newt — Creature — Nightmare Salamander Beast {1}{U}, 2/2 (Judgment):
//
//	"When this creature enters, exile a land you control.
//	 When this creature leaves the battlefield, return the exiled card to the
//	 battlefield under its owner's control."
//
// The linked pair of S82OldB626ExileALandPair, shared with Wormfang Turtle.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "be014a79-5550-4f79-8789-b4d6783d7ec8",
		Name:         "Wormfang Newt",
		Completeness: effects.CompletenessFull,
		Triggered:    effects.S82OldB626ExileALandPair("Wormfang Newt", "choose-cards/s82-old-b626-wormfang-newt-land"),
	})
}
