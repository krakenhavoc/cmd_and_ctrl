package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Culling Scales — Artifact {3} (Mirrodin, #5606):
//
//	"At the beginning of your upkeep, destroy target nonland permanent
//	 with the lowest mana value. (If two or more permanents are tied for
//	 lowest, target any one of them.)"
//
// A mandatory targeted upkeep trigger. The target clause is a nonland
// permanent no other nonland permanent undercuts on mana value
// (LowestManaValueNonland), anyone's, the Scales themselves included:
// with nothing cheaper on the battlefield, the Scales are the target. A
// tie is any one of the tied permanents. The clause is judged again as
// the trigger resolves, so a cheaper permanent arriving in response
// leaves the trigger with nothing to do (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ec4f2266-3e8e-4194-8f06-d95ebc543661",
		Name:         "Culling Scales",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtYourUpkeep("Culling Scales — destroy target nonland permanent with the lowest mana value", destroyFirstLegalCardTarget),
				TargetPermanent("target nonland permanent with the lowest mana value", LowestManaValueNonland())),
		},
	})
}
