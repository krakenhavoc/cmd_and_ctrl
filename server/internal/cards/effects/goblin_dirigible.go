package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Dirigible — Artifact Creature — Construct {6}, 4/4 (Mirrodin,
// #5606):
//
//	"Flying
//	 This creature doesn't untap during your untap step.
//	 At the beginning of your upkeep, you may pay {4}. If you do, untap
//	 this creature."
//
// Brass Man at {4}, with flying: the catalog's untap-step restriction
// (CR 502.3) and the upkeep "you may pay" (upkeepMayPayToUntapThis).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:              "8e69b322-4e66-4d5b-ab46-f002c40a913f",
		Name:                  "Goblin Dirigible",
		Completeness:          CompletenessFull,
		PrintedKeywords:       []string{"flying"},
		UntapStepRestrictions: []game.UntapStepRestriction{doesntUntapDuringYourUntapStep()},
		Triggered: []game.TriggeredAbility{
			upkeepMayPayToUntapThis("Goblin Dirigible", "{4}", goblinDirigiblePaid),
		},
	})
}

// goblinDirigiblePaid is the paid half, keyed (#4264). An on-disk key:
// never renamed, never reused.
var goblinDirigiblePaid = PayUnlessBranches("pay-unless/goblin-dirigible-untap", untapThisIfStillOnBattlefield, nil)
