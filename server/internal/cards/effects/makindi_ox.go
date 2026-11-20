package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Makindi Ox — Creature — Ox {4}{W}, 4/4 (Zendikar Rising, #5508):
//
//	"Landfall — Whenever a land you control enters, tap target creature
//	 an opponent controls."
//
// Landfall is an ability word (CR 207.2c). The tap is Captivating
// Unicorn's: the target is chosen as the trigger goes on the stack, and a
// target that left or stopped being legal does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e4f0fa12-542b-485f-a792-3a8aba933522",
		Name:         "Makindi Ox",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				Landfall("Makindi Ox — tap target creature an opponent controls (landfall)", tapChosenPermanent),
				TargetCreature("target creature an opponent controls", OpponentControls()),
			),
		},
	})
}
