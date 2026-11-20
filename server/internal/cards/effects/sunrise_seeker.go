package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sunrise Seeker — Creature — Human Scout {4}{W}, 3/3 (Ixalan, #5532):
//
//	"Vigilance
//	 When this creature enters, it explores."
//
// Vigilance is an engine keyword carried as printed data: attacking
// doesn't tap it (CR 702.20b).
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
		OracleID:        "db05f30f-2866-48ed-8a7d-264d94c0257a",
		Name:            "Sunrise Seeker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sunrise Seeker — it explores", thatCreatureExplores),
		},
	})
}
