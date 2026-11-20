package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Akoum Hellhound — Creature — Elemental Dog {R}, 0/1 (Zendikar Rising,
// #5508):
//
//	"Landfall — Whenever a land you control enters, this creature gets
//	 +2/+2 until end of turn."
//
// Saurian Explorer's landfall pump (an ability word, CR 207.2c): one
// trigger per land, each +2/+2 until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "05074b68-6b31-4f9f-b3ec-bffe3e7eab47",
		Name:         "Akoum Hellhound",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Landfall("Akoum Hellhound — +2/+2 until end of turn (landfall)",
				thisGetsUntilEndOfTurn(2, 2, "Akoum Hellhound — +2/+2 until end of turn")),
		},
	})
}
