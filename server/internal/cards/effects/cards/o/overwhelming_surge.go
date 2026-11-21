package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Overwhelming Surge — Instant {2}{R} (Tarkir: Dragonstorm, #4463):
//
//	"Choose one or both —
//	 • Overwhelming Surge deals 3 damage to target creature.
//	 • Destroy target noncreature artifact."
//
// Deadly Complication's "Choose one or both" (ChooseN 1..2). Each bullet
// has its own target clause, so a bullet with no legal target can't be
// chosen (CR 700.2a), and with both chosen the bullets run in printed
// order: the damage, then the destruction. A bullet whose target became
// illegal does nothing while the other still happens (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bc8a37f1-c36b-4d0f-82da-4558ff712565",
		Name:         "Overwhelming Surge",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose one or both", 1, 2,
			effects.ModeWithPurpose(effects.ModeDoing("Overwhelming Surge deals 3 damage to target creature.",
				effects.TargetCreature("target creature"), effects.DealFixedDamageToModesTarget(3)),
				effects.ForTargets(effects.DamageToTarget(0, 3))),
			effects.ModeDoing("Destroy target noncreature artifact.",
				effects.TargetPermanent("target noncreature artifact", effects.Artifact(), effects.Noncreature()), effects.DestroyTheModesTarget),
		),
	})
}
