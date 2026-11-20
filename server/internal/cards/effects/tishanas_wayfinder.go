package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tishana's Wayfinder — Creature — Merfolk Scout {2}{G}, 2/2 (Ixalan, #5532):
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
		OracleID:     "86468528-833d-48f9-b299-cdf54dd47d77",
		Name:         "Tishana's Wayfinder",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Tishana's Wayfinder — it explores", thatCreatureExplores),
		},
	})
}
