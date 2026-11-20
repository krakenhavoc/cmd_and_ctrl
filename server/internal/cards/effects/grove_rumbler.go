package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grove Rumbler — Creature — Elemental {2}{R}{G}, 3/3 (Battle for
// Zendikar, #5547):
//
//	"Trample
//	 Landfall — Whenever a land you control enters, this creature
//	 gets +2/+2 until end of turn."
//
// Canopy Baloth's landfall on a trampler. Landfall is the stack trigger
// on a land entering under its controller's control (CR 207.2c), so a
// land an opponent plays does nothing; the pump ends in the cleanup
// step (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5dcbed75-bafc-492a-85a1-44f3ff06fa78",
		Name:            "Grove Rumbler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			Landfall("Grove Rumbler — +2/+2 until end of turn (landfall)",
				thisGetsUntilEndOfTurn(2, 2, "Grove Rumbler — +2/+2 until end of turn")),
		},
	})
}
