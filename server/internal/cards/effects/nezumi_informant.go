package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nezumi Informant — Creature — Rat Rogue 1/1 {1}{B} (March of the Machine, S82):
//
//	"When this creature enters, each opponent discards a card."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a70ffefb-840c-48c3-a59d-e12306086bb6",
		Name:         "Nezumi Informant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Nezumi Informant — each opponent discards a card", eachOpponentDiscardsOne),
		},
	})
}
