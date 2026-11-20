package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Weed-Pruner Poplar — Creature — Treefolk Assassin {4}{B}, 3/3
// (Morningtide, #5592):
//
//	"At the beginning of your upkeep, target creature other than this
//	 creature gets -1/-1 until end of turn."
//
// A mandatory upkeep trigger whose target clause excludes the source
// object (Another, CR 109.1 identity, not a name); any other creature is
// legal, yours included, and with none the trigger is removed for want of
// a target (CR 603.3d). The body is boostTargetUntilEOT with -1/-1.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "40e8ade9-9cec-4983-ab11-d9e6278fb29f",
		Name:         "Weed-Pruner Poplar",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(AtYourUpkeep("Weed-Pruner Poplar — target creature other than this creature gets -1/-1",
				boostTargetUntilEOT("Weed-Pruner Poplar — -1/-1", -1, -1)),
				Another(TargetCreature("target creature other than this creature"))),
				ForTargets(HarmUntilEOT(0, -1, -1))),
		},
	})
}
