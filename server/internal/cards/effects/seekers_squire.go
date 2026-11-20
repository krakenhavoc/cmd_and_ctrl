package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seekers' Squire — Creature — Human Scout {1}{B}, 1/2 (Ixalan, #5532):
//
//	"When this creature enters, it explores."
//
// The explorer is the object that entered, read off the trigger
// (thatCreatureExplores, Cenote Scout's body): a land on top goes to
// the hand, anything else gives this creature a +1/+1 counter and asks
// whether to put the card into the graveyard (CR 701.44a). One that has
// left by the time the trigger resolves still explores by its last
// known information: the card is revealed, and no counter is put on
// anything (CR 701.44c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a494cbe7-a4a5-4657-9f68-3811fe56a17e",
		Name:         "Seekers' Squire",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Seekers' Squire — it explores", thatCreatureExplores),
		},
	})
}
