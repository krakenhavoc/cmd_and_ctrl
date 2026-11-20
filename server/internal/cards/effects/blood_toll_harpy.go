package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blood-Toll Harpy — Creature — Harpy {2}{B}, 2/1 (Theros, #5561):
//
//	"Flying
//	 When this creature enters, each player loses 1 life."
//
// The enters trigger takes 1 life from every player at the table, its
// controller included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "98f9619e-585f-486f-b143-fb42085eb583",
		Name:            "Blood-Toll Harpy",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Blood-Toll Harpy — each player loses 1 life", Do(Instruction(socB22EachPlayerLosesOneLife))),
		},
	})
}
