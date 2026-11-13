package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin War Wagon — Artifact Creature — Juggernaut {4}, 3/3 (Mirrodin,
// #5606):
//
//	"This creature doesn't untap during your untap step.
//	 At the beginning of your upkeep, you may pay {2}. If you do, untap
//	 this creature."
//
// Goblin Dirigible's two lines at {2}, without flying: the untap-step
// restriction (CR 502.3) and the upkeep "you may pay"
// (upkeepMayPayToUntapThis).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:              "81a38340-9558-49d3-8de6-d7d01babfd97",
		Name:                  "Goblin War Wagon",
		Completeness:          CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{doesntUntapDuringYourUntapStep()},
		Triggered: []game.TriggeredAbility{
			upkeepMayPayToUntapThis("Goblin War Wagon", "{2}", goblinWarWagonPaid),
		},
	})
}

// goblinWarWagonPaid is the paid half, keyed (#4264). An on-disk key:
// never renamed, never reused.
var goblinWarWagonPaid = PayUnlessBranches("pay-unless/goblin-war-wagon-untap", untapThisIfStillOnBattlefield, nil)
