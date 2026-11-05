package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kyoshi Warriors — Creature — Human Warrior Ally {3}{W}, 3/3:
//
//	"When this creature enters, create a 1/1 white Ally creature
//	 token."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b29abb09-75ea-431c-ab60-f2940d002e83",
		Name:         "Kyoshi Warriors",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Kyoshi Warriors — create a 1/1 Ally", Do(CreateToken{Template: TokenCard("1/1 white Ally"), N: 1})),
		},
	})
}
