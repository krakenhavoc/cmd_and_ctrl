package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Destroy Evil — Instant {1}{W}:
//
//	"Choose one —
//	 • Destroy target creature with toughness 4 or greater.
//	 • Destroy target enchantment."
//
// Two destroy bullets on the shared DestroyTheModesTarget body. The
// toughness is judged at announce and again at resolution (CR 608.2b),
// so a creature shrunk below 4 in response is no longer a legal target
// and the spell does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b75e667a-1ceb-445a-acf4-ae1fd2dae9e7",
		Name:         "Destroy Evil",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Destroy target creature with toughness 4 or greater.",
				effects.TargetCreature("target creature with toughness 4 or greater", effects.ToughnessGE(4)),
				effects.DestroyTheModesTarget),
			effects.ModeDoing("Destroy target enchantment.",
				effects.TargetPermanent("target enchantment", effects.Enchantment()),
				effects.DestroyTheModesTarget),
		),
	})
}
