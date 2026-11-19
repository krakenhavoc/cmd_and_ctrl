package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tukatongue Thallid — Creature — Fungus {G}, 1/1 (Foundations Jumpstart):
//
//	"When this creature dies, create a 1/1 green Saproling creature token."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c7dbd008-e941-497e-8305-80c53b1dffe2",
		Name:         "Tukatongue Thallid",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(WhenThisDies("Tukatongue Thallid — create a 1/1 green Saproling creature token",
				Do(CreateToken{Template: TokenCard("1/1 green Saproling"), N: 1})), game.Purpose{Tokens: 1}),
		},
	})
}
