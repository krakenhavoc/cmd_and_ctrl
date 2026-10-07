package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Instant Ramen — Artifact — Food {2}:
//
//	"Flash
//	 When this artifact enters, draw a card.
//	 {2}, {T}, Sacrifice this artifact: You gain 3 life."
//
// A Food that cantrips and can be cast at instant speed. No
// simplification.
func init() {
	Register(Spec{
		OracleID:        "2283e409-c6c7-4de9-899b-2b3caea5f35e",
		Name:            "Instant Ramen",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{samiDrawOnETB("Instant Ramen")},
		Activated:       []ActivatedAbility{samiFoodLifeAbility()},
	})
}
