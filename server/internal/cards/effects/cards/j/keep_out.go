package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Keep Out — Instant {1}{W} (Lorwyn Eclipsed, #4458):
//
//	"Choose one —
//	 • Keep Out deals 4 damage to target tapped creature.
//	 • Destroy target enchantment."
//
// Two bullets, each with its own target clause, judged as the spell
// resolves (CR 608.2b): a creature untapped in response is an illegal
// target and the damage does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "61cb927a-4cf8-49d6-ab4b-9e882bc08774",
		Name:         "Keep Out",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Keep Out deals 4 damage to target tapped creature.",
				effects.TargetCreature("target tapped creature", effects.TappedPermanent()),
				effects.DealFixedDamageToModesTarget(4)),
			effects.ModeDoing("Destroy target enchantment.",
				effects.TargetPermanent("target enchantment", effects.Enchantment()),
				effects.DestroyTheModesTarget),
		),
	})
}
