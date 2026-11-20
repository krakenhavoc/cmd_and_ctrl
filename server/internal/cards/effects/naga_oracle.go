package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Naga Oracle — Creature — Snake Cleric {3}{U}, 2/4
// (Amonkhet, #5535):
//
//	"When this creature enters, surveil 3. (Look at the top three cards of your library, then put any number of them into your graveyard and the rest on top of your library in any order.)"
//
// Cephalid Inkmage's enters trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "249e18a4-12c5-4836-a172-cb4e9b9cb7d7",
		Name:         "Naga Oracle",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Naga Oracle — surveil 3", Do(Surveil{N: 3})),
		},
	})
}
