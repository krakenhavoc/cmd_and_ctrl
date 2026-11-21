package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deathbloom Thallid — Creature — Fungus {2}{B}, 3/2 (Dominaria, #5530):
//
//	"When this creature dies, create a 1/1 green Saproling creature token."
//
// Tukatongue Thallid's dies trigger: one 1/1 green Saproling.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ce873790-de4d-4ba3-8023-c44306ba3ff8",
		Name:         "Deathbloom Thallid",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WhenThisDies("Deathbloom Thallid — create a 1/1 green Saproling creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Saproling"), N: 1})), game.Purpose{Tokens: 1}),
		},
	})
}
