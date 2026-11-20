package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Borderland Marauder — Creature — Human Warrior {1}{R}, 1/2 (Magic 2015,
// #5556):
//
//	"Whenever this creature attacks, it gets +2/+0 until end of turn."
//
// Charging Bandits' attack trigger, word for word.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "11fd2650-2150-4e62-8b64-e820d734a231",
		Name:         "Borderland Marauder",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Borderland Marauder — it gets +2/+0 until end of turn",
				thisGetsUntilEndOfTurn(2, 0, "Borderland Marauder — +2/+0")),
		},
	})
}
