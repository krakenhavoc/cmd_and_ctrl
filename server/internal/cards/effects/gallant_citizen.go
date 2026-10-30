package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gallant Citizen — Creature — Human Citizen {G/W}{G/W}, 1/1 (Marvel
// Super Heroes Commander, #4455): "When this creature enters, draw a
// card." The same shape as Angelic Gift's ETB draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2fcb1c1c-6157-4955-b899-e8a8b337d188",
		Name:         "Gallant Citizen",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Gallant Citizen — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
