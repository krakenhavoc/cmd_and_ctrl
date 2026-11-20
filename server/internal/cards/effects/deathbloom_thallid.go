package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deathbloom Thallid — Creature — Fungus {2}{B}, 3/2 (Dominaria, #5530):
//
//	"When this creature dies, create a 1/1 green Saproling creature token."
//
// Tukatongue Thallid's dies trigger: one 1/1 green Saproling.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ce873790-de4d-4ba3-8023-c44306ba3ff8",
		Name:         "Deathbloom Thallid",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(WhenThisDies("Deathbloom Thallid — create a 1/1 green Saproling creature token",
				Do(CreateToken{Template: TokenCard("1/1 green Saproling"), N: 1})), game.Purpose{Tokens: 1}),
		},
	})
}
