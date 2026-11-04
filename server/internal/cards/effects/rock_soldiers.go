package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rock Soldiers — {3}{R} Artifact Creature — Elemental Soldier 4/3 (Teenage
// Mutant Ninja Turtles, #4457):
//
//	"When this creature enters, destroy up to one target noncreature
//	 artifact."
//
// Storm, Shaker of Skies's "destroy up to one target" enters trigger with a
// noncreature artifact clause. Choosing no target is legal.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0217da8e-e74f-4f5c-ab75-f17600f94405",
		Name:         "Rock Soldiers",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Rock Soldiers — destroy up to one target noncreature artifact", destroyFirstLegalTarget),
				TargetPermanent("up to one target noncreature artifact", And(Artifact(), Noncreature())).WithCount(0, 1)),
		},
	})
}
