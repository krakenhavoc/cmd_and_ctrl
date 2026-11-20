package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kiln Walker — Artifact Creature — Phyrexian Construct {3}, 0/3 (New Phyrexia, #5575):
//
//	"Whenever this creature attacks, it gets +3/+0 until end of turn."
//
// Charging Bandits' attack trigger at +3/+0.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ca868ac4-6144-402e-ae23-629628d02883",
		Name:         "Kiln Walker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Kiln Walker — it gets +3/+0 until end of turn",
				thisGetsUntilEndOfTurn(3, 0, "Kiln Walker — +3/+0")),
		},
	})
}
