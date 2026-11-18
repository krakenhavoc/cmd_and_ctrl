package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Toxic Scorpion — Creature — Scorpion {1}{G}, 1/1 (Innistrad: Crimson
// Vow):
//
//	"Deathtouch
//	 When this creature enters, another target creature you control gains
//	 deathtouch until end of turn."
//
// "Another" excludes the Scorpion itself (object identity). The target is
// re-checked as the trigger resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "612ea70d-f4f2-4909-a025-218161ae1059",
		Name:            "Toxic Scorpion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Toxic Scorpion — another target creature you control gains deathtouch",
				ebTargetCreatureGainsUntilEOT("deathtouch", "Toxic Scorpion — deathtouch")),
				Another(TargetCreature("another target creature you control", YouControl()))),
				ForTargets(BenefitUntilEOT(0, 0, 0, "deathtouch"))),
		},
	})
}
