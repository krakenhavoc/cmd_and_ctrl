package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Countless Gears Renegade — Creature — Dwarf Artificer {1}{W}, 2/2:
//
//	"Revolt — When this creature enters, if a permanent left the
//	 battlefield under your control this turn, create a 1/1 colorless
//	 Servo artifact creature token."
//
// Revolt is an intervening if (revolt.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9d1f9036-a466-46b6-ab7c-2de487876978",
		Name:         "Countless Gears Renegade",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersIfRevolt("Countless Gears Renegade — create a 1/1 colorless Servo artifact creature token",
				Do(CreateToken{Template: TokenCard("1/1 colorless Servo artifact"), N: 1})),
		},
	})
}
