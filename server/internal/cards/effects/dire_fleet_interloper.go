package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dire Fleet Interloper — Creature — Human Pirate {3}{B}, 2/2 (Ixalan,
// #5532):
//
//	"Menace
//	 When this creature enters, it explores."
//
// Menace is an engine keyword carried as printed data: it can't be
// blocked except by two or more creatures (CR 702.111b).
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
		OracleID:        "7ae7cfde-8f2a-4e66-abe2-c15477e9fac3",
		Name:            "Dire Fleet Interloper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Dire Fleet Interloper — it explores", thatCreatureExplores),
		},
	})
}
