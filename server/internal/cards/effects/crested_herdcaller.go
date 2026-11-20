package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crested Herdcaller — Creature — Dinosaur {3}{G}{G}, 3/3 (Rivals of
// Ixalan, #5531):
//
//	"Trample
//	 When this creature enters, create a 3/3 green Dinosaur creature token
//	 with trample."
//
// Trample is printed data (CR 702.19b); the token is Thundering
// Spineback's 3/3 Dinosaur from the token table.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "86c0ddaa-6926-450d-949a-6c89f43c71c2",
		Name:            "Crested Herdcaller",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Crested Herdcaller — create a 3/3 green Dinosaur with trample",
				Do(CreateToken{Template: TokenCard("3/3 green Dinosaur with trample"), N: 1})),
		},
	})
}
