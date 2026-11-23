package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cleansing Ray — Sorcery {1}{W} (Rivals of Ixalan, #5531):
//
//	"Choose one —
//	 • Destroy target Vampire.
//	 • Destroy target enchantment."
//
// Each bullet's target clause is its own, so the picker offers Vampires
// or enchantments by mode; the body is the shared DestroyTheModesTarget.
// "Vampire" is any permanent of that type, not only a creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "50f1948c-fa27-458d-be14-cb0e09e39f63",
		Name:         "Cleansing Ray",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Destroy target Vampire.", effects.TargetPermanent("target Vampire", effects.Subtype("Vampire")), effects.DestroyTheModesTarget),
			effects.ModeDoing("Destroy target enchantment.", effects.TargetPermanent("target enchantment", effects.Enchantment()), effects.DestroyTheModesTarget),
		),
	})
}
