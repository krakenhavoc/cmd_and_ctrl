package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grave Defiler — Creature — Zombie {3}{B}, 2/1 (Apocalypse, #5614):
//
//	"When this creature enters, reveal the top four cards of your library.
//	 Put all Zombie cards revealed this way into your hand and the rest on the
//	 bottom of your library in any order.
//	 {1}{B}: Regenerate this creature."
//
// Sylvan Messenger's reveal (the Zombie word is this card's text): every Zombie
// card among the four is taken, none is chosen, and the controller orders
// the rest on the bottom. A library shorter than four reveals what it has.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "42c5c1aa-10a4-4cd5-b1f8-a393ed553f99",
		Name:         "Grave Defiler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Grave Defiler — reveal the top four cards; Zombie cards to your hand, the rest on the bottom in any order",
				effects.S82OldB634RevealFourTakeSubtype("Grave Defiler", "Zombie")),
		},
		Activated: []effects.ActivatedAbility{effects.LrwB1RegenerateThisRow("{1}{B}")},
	})
}
