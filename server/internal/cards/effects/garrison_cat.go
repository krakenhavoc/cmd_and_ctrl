package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Garrison Cat — Creature — Cat {W}, 1/1 (Ikoria: Lair of Behemoths, #5512):
//
//	"When this creature dies, create a 1/1 white Human Soldier creature
//	 token."
//
// Doomed Traveler's death trigger with a Human Soldier.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e7bc7c44-946c-4131-b764-add0c5f21020",
		Name:         "Garrison Cat",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Garrison Cat — create a 1/1 white Human Soldier creature token",
				Do(CreateToken{Template: TokenCard("1/1 white Human Soldier"), N: 1})),
		},
	})
}
