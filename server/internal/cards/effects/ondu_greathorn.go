package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ondu Greathorn — Creature — Beast {3}{W}, 2/3 (Battle for Zendikar,
// #5547):
//
//	"First strike
//	 Landfall — Whenever a land you control enters, this creature
//	 gets +2/+2 until end of turn."
//
// Canopy Baloth's landfall on a first striker. Landfall is the stack
// trigger on a land entering under its controller's control (CR
// 207.2c), so a land an opponent plays does nothing; the pump ends in
// the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9d061ec5-1edf-46e7-9e6a-4ddc9ced4114",
		Name:            "Ondu Greathorn",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			Landfall("Ondu Greathorn — +2/+2 until end of turn (landfall)",
				thisGetsUntilEndOfTurn(2, 2, "Ondu Greathorn — +2/+2 until end of turn")),
		},
	})
}
