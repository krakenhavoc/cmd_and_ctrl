package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Enlistment Officer — Creature — Human Soldier {3}{W}, 2/3 (Apocalypse, #5614):
//
//	"First strike
//	 When this creature enters, reveal the top four cards of your library.
//	 Put all Soldier cards revealed this way into your hand and the rest on the
//	 bottom of your library in any order."
//
// Sylvan Messenger's reveal (the Soldier word is this card's text): every Soldier
// card among the four is taken, none is chosen, and the controller orders
// the rest on the bottom. A library shorter than four reveals what it has.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e9517fc3-d1bc-42f7-b052-9f629b302735",
		Name:            "Enlistment Officer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Enlistment Officer — reveal the top four cards; Soldier cards to your hand, the rest on the bottom in any order",
				effects.S82OldB634RevealFourTakeSubtype("Enlistment Officer", "Soldier")),
		},
	})
}
