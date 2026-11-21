package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Yavimaya Sapherd — Creature — Fungus {2}{G}, 2/2 (Dominaria, #5530):
//
//	"When this creature enters, create a 1/1 green Saproling creature token."
//
// One 1/1 green Saproling when it enters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "42a1954f-774e-4d80-98b3-a736f23e2060",
		Name:         "Yavimaya Sapherd",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WhenThisEnters("Yavimaya Sapherd — create a 1/1 green Saproling creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Saproling"), N: 1})), game.Purpose{Tokens: 1}),
		},
	})
}
