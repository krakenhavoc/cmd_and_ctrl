package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wormfang Turtle — Creature — Nightmare Turtle Beast {2}{U}, 2/4 (Judgment):
//
//	"When this creature enters, exile a land you control.
//	 When this creature leaves the battlefield, return the exiled card to the
//	 battlefield under its owner's control."
//
// The linked pair of S82OldB626ExileALandPair, shared with Wormfang Newt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "771f9ba4-6c65-4497-ad48-869994435556",
		Name:         "Wormfang Turtle",
		Completeness: effects.CompletenessFull,
		Triggered:    effects.S82OldB626ExileALandPair("Wormfang Turtle", "choose-cards/s82-old-b626-wormfang-turtle-land"),
	})
}
