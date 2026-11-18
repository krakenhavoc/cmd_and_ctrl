package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ritual Guardian — Creature — Human Soldier {2}{W}, 3/2 (Midnight Hunt,
// #5497):
//
//	"Coven — At the beginning of combat on your turn, if you control
//	 three or more creatures with different powers, this creature gains
//	 lifelink until end of turn."
//
// Candlegrove Witch's trigger with lifelink (midB05CovenThisGainsKeyword).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9f4719f2-0eab-464b-9ae5-2b96717ae08d",
		Name:         "Ritual Guardian",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{midB05CovenThisGainsKeyword("Ritual Guardian", "lifelink")},
	})
}
