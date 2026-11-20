package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Siren Lookout — Creature — Siren Pirate {2}{U}, 1/2 (Ixalan, #5532):
//
//	"Flying
//	 When this creature enters, it explores."
//
// Flying is an engine keyword carried as printed data: it can be
// blocked only by creatures with flying or reach (CR 702.9b).
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
		OracleID:        "9f0d82ae-38bf-45d8-8cda-982b6ead1d72",
		Name:            "Siren Lookout",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Siren Lookout — it explores", thatCreatureExplores),
		},
	})
}
