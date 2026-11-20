package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Makindi Sliderunner — Creature — Beast {1}{R}, 2/1 (Battle for
// Zendikar, #5547):
//
//	"Trample
//	 Landfall — Whenever a land you control enters, this creature
//	 gets +1/+1 until end of turn."
//
// Snapping Gnarlid's landfall on a trampler. Landfall is the stack
// trigger on a land entering under its controller's control (CR
// 207.2c), so a land an opponent plays does nothing; the pump ends in
// the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4754e9c6-1790-4c01-87c2-3a16ce7e65f5",
		Name:            "Makindi Sliderunner",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			Landfall("Makindi Sliderunner — +1/+1 until end of turn (landfall)",
				thisGetsUntilEndOfTurn(1, 1, "Makindi Sliderunner — +1/+1 until end of turn")),
		},
	})
}
