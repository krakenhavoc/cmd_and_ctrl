package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "612ea70d-f4f2-4909-a025-218161ae1059",
		Name:            "Toxic Scorpion",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Toxic Scorpion — another target creature you control gains deathtouch",
				effects.EbTargetCreatureGainsUntilEOT("deathtouch", "Toxic Scorpion — deathtouch")),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
				effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "deathtouch"))),
		},
	})
}
