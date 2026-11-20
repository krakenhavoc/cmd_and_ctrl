package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scythe Leopard — Creature — Cat {G}, 1/1 (Battle for Zendikar,
// #5547):
//
//	"Landfall — Whenever a land you control enters, this creature
//	 gets +1/+1 until end of turn."
//
// Snapping Gnarlid's landfall. Landfall is the stack trigger on a land
// entering under its controller's control (CR 207.2c), so a land an
// opponent plays does nothing; the pump ends in the cleanup step (CR
// 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ab539b16-d2a0-404d-8ad3-7ff33f2f1abe",
		Name:         "Scythe Leopard",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Landfall("Scythe Leopard — +1/+1 until end of turn (landfall)",
				thisGetsUntilEndOfTurn(1, 1, "Scythe Leopard — +1/+1 until end of turn")),
		},
	})
}
