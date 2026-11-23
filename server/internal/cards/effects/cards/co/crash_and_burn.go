package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crash and Burn — Instant {3}{R} (Aetherdrift, #4464):
//
//	"Choose one —
//	 • Destroy target Vehicle.
//	 • Crash and Burn deals 6 damage to target creature or planeswalker."
//
// Destroy Evil's modal shape. The first bullet's Vehicle may be crewed or
// not; the second is Fulminous Forte's fixed damage to a creature or
// planeswalker. Each bullet's target is re-checked as it resolves
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "38405a37-921a-49ab-a95b-28eb8c26474a",
		Name:         "Crash and Burn",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Destroy target Vehicle.",
				effects.TargetPermanent("target Vehicle", effects.HasSubtype("Vehicle")),
				effects.DestroyTheModesTarget),
			effects.ModeWithPurpose(effects.ModeDoing("Crash and Burn deals 6 damage to target creature or planeswalker.",
				effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
				effects.DealFixedDamageToModesTarget(6)), effects.ForTargets(effects.DamageToTarget(0, 6))),
		),
	})
}
