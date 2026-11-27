package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tidal Courier — Creature — Merfolk {3}{U}, 1/2 (Apocalypse, #5614):
//
//	"When this creature enters, reveal the top four cards of your library.
//	 Put all Merfolk cards revealed this way into your hand and the rest on the
//	 bottom of your library in any order.
//	 {3}{U}: This creature gains flying until end of turn."
//
// Sylvan Messenger's reveal (the Merfolk word is this card's text): every Merfolk
// card among the four is taken, none is chosen, and the controller orders
// the rest on the bottom. A library shorter than four reveals what it has.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b0a0c17b-87c8-4610-adca-5e7f5ced49f6",
		Name:         "Tidal Courier",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Tidal Courier — reveal the top four cards; Merfolk cards to your hand, the rest on the bottom in any order",
				effects.S82OldB634RevealFourTakeSubtype("Tidal Courier", "Merfolk")),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{U}: This creature gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.ManaCost("{3}{U}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("flying", "Tidal Courier — flying"),
		}},
	})
}
