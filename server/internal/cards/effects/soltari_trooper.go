package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soltari Trooper — Creature — Soltari Soldier {1}{W}, 1/1
// (Tempest, #5628):
//
//	"Shadow (This creature can block or be blocked by only creatures
//	 with shadow.)
//	 Whenever this creature attacks, it gets +1/+1 until end of turn."
//
// Shadow is a printed keyword. The attack trigger is the shared
// WheneverThisAttacks shape (CR 508.1m) with the thisGetsUntilEndOfTurn
// body, so the Trooper attacks as a 2/2, and the boost ends in the
// cleanup step (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "501109bb-3406-41d8-9433-a100e494f938",
		Name:            "Soltari Trooper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"shadow"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Soltari Trooper — it gets +1/+1 until end of turn",
				thisGetsUntilEndOfTurn(1, 1, "Soltari Trooper — +1/+1")),
		},
	})
}
