package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Steadfast Cathar — Creature — Human Soldier {1}{W}, 2/1 (Eldritch Moon,
// #5541):
//
//	"Whenever this creature attacks, it gets +0/+2 until end of turn."
//
// Charging Paladin's attack trigger with +0/+2.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "00895741-f94f-4d80-8ca4-c436bf3c8d56",
		Name:         "Steadfast Cathar",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Steadfast Cathar — it gets +0/+2 until end of turn",
				thisGetsUntilEndOfTurn(0, 2, "Steadfast Cathar — +0/+2")),
		},
	})
}
