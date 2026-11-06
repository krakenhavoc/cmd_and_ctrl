package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cenote Scout — Creature — Merfolk Scout {G}, 1/1 (The Lost Caverns of
// Ixalan, #4470):
//
//	"When this creature enters, it explores."
//
// The explorer is the object that entered, read off the trigger
// (thatCreatureExplores, Path of Discovery's body): a land on top goes to
// the hand, anything else gives the Scout a +1/+1 counter and asks
// whether to put the card into the graveyard (CR 701.44a). A Scout that
// has left by the time the trigger resolves still explores by its last
// known information: the card is revealed, and no counter is put on
// anything (CR 701.44c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ec3911c0-686f-4c78-838d-a683eed6c01c",
		Name:         "Cenote Scout",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Cenote Scout — it explores", thatCreatureExplores),
		},
	})
}
