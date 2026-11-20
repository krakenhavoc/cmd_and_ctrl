package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Watchful Giant — Creature — Giant Soldier {5}{W}, 3/6 (Ravnica Allegiance,
// #5521):
//
//	"When this creature enters, create a 1/1 white Human creature token."
//
// Clarion Cathars's enters trigger and token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8239c63d-097a-449a-953a-cd480e594aa8",
		Name:         "Watchful Giant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Watchful Giant — create a 1/1 white Human creature token",
				Do(CreateToken{Template: TokenCard("1/1 white Human"), N: 1})),
		},
	})
}
