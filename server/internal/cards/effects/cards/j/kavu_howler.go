package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kavu Howler — Creature — Kavu {4}{G}{G}, 4/5 (Apocalypse, #5614):
//
//	"When this creature enters, reveal the top four cards of your library.
//	 Put all Kavu cards revealed this way into your hand and the rest on the
//	 bottom of your library in any order."
//
// Sylvan Messenger's reveal (the Kavu word is this card's text): every Kavu
// card among the four is taken, none is chosen, and the controller orders
// the rest on the bottom. A library shorter than four reveals what it has.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "635c0a7d-d832-4c10-a5f0-9cde19152cbb",
		Name:         "Kavu Howler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Kavu Howler — reveal the top four cards; Kavu cards to your hand, the rest on the bottom in any order",
				effects.S82OldB634RevealFourTakeSubtype("Kavu Howler", "Kavu")),
		},
	})
}
