package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kinjalli's Dawnrunner — Creature — Human Scout {2}{W}, 1/1 (The Lost
// Caverns of Ixalan, #4470):
//
//	"Double strike
//	 When this creature enters, it explores."
//
// Cenote Scout's enters trigger (thatCreatureExplores) on a printed
// double striker: a land on top goes to the hand, anything else gives the
// Dawnrunner a +1/+1 counter and asks whether the card goes to the
// graveyard (CR 701.44a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b6565da9-773c-49b1-85f5-e34f07ba84cb",
		Name:            "Kinjalli's Dawnrunner",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Kinjalli's Dawnrunner — it explores", thatCreatureExplores),
		},
	})
}
