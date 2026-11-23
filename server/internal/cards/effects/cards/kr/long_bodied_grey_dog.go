package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Long-Bodied Grey Dog — Creature — Dog {3}, 2/2 (The Hobbit, HOB):
//
//	"Flash
//	 Reach
//	 When this creature enters, create a tapped Treasure token."
//
// Both keywords ride PrintedKeywords; the trigger creates one Treasure
// that enters tapped, as Nuka-Cola Vending Machine's does (#4454).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6f83da19-fd89-44ec-88f3-0c3fddfbd1b2",
		Name:            "Long-Bodied Grey Dog",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "reach"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Long-Bodied Grey Dog — create a tapped Treasure token",
				effects.Do(effects.CreateTokenAdvanced{Spec: effects.Token(effects.TreasureToken()).EntersTapped(), N: 1})),
		},
	})
}
