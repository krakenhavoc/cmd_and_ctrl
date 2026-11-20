package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sanguine Glorifier — Creature — Vampire Cleric {3}{W}, 3/3 (Rivals of
// Ixalan, #5531):
//
//	"When this creature enters, put a +1/+1 counter on another target
//	 Vampire you control."
//
// Jade Bearer's entry for Vampires: "another" is object identity, so the
// Glorifier can't pick itself, and with no other Vampire of yours the
// trigger is removed for want of a target (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "391f1331-7577-4bf4-a9ed-2f55ded4cbd8",
		Name:         "Sanguine Glorifier",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Sanguine Glorifier — put a +1/+1 counter on another target Vampire you control",
				rfCreatureFCountersOnTarget(1)),
				Another(TargetCreature("another target Vampire you control", YouControl(), Subtype("Vampire")))),
		},
	})
}
