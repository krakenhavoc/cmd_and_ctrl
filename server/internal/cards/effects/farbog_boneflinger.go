package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Farbog Boneflinger — Creature — Zombie {4}{B}, 2/2 (Dark Ascension, #5571):
//
//	"When this creature enters, target creature gets -2/-2 until end
//	 of turn."
//
// A layer 7c -2/-2 on the still-legal target (CR 608.2b), gone at cleanup.
// The target is any creature, the Boneflinger included when it is the only
// one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "46027ca0-9d06-42b3-8012-124861c6b558",
		Name:         "Farbog Boneflinger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Farbog Boneflinger — target creature gets -2/-2 until end of turn",
				boostTargetUntilEOT("Farbog Boneflinger — -2/-2", -2, -2)),
				TargetCreature("target creature")), ForTargets(HarmUntilEOT(0, -2, -2))),
		},
	})
}
