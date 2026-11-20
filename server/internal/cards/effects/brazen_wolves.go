package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brazen Wolves — Creature — Wolf {2}{R}, 2/3 (Eldritch Moon, #5541):
//
//	"Whenever this creature attacks, it gets +2/+0 until end of turn."
//
// Charging Bandits' attack trigger, word for word.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0b5a3d7b-6b2c-4e91-8218-2b957277855c",
		Name:         "Brazen Wolves",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Brazen Wolves — it gets +2/+0 until end of turn",
				thisGetsUntilEndOfTurn(2, 0, "Brazen Wolves — +2/+0")),
		},
	})
}
