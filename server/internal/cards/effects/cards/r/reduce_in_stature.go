package cardsr

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// Reduce in Stature — Enchantment — Aura {2}{U} (Dragons of Tarkir, S82 old batch
// 332):
//
//	"Enchant creature
//	 Enchanted creature has base power and toughness 0/2."
//
// Illusory Wrappings' layer 7b set, through the shared RegisterBasePTAura: the
// base is set, so counters and other pumps still apply on top (CR 613.4b,
// 613.4c).
//
// No simplification.
func init() {
	effects.RegisterBasePTAura("7ebaeef0-4e4c-4bf0-85dd-07c6b6be0741", "Reduce in Stature", 0, 2)
}
