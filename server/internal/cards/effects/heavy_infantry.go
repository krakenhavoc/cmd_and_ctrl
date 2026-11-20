package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Heavy Infantry — Creature — Human Soldier {4}{W}, 3/4 (Magic Origins, #5549):
//
//	"When this creature enters, tap target creature an opponent controls."
//
// Vexing Drone's entry: the opponent's creature is targeted as the
// trigger goes on the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d07707ef-fa4e-423f-8a1f-095d19100321",
		Name:         "Heavy Infantry",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Heavy Infantry — tap target creature an opponent controls", tapTheFirstLegalTarget),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
