package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wind Zendikon — Enchantment — Aura {U} (Worldwake, S82 OLD batch 489, #5582):
//
//	"Enchant land
//	 Enchanted land is a 2/2 blue Elemental creature with flying. It's still a land.
//	 When enchanted land dies, return that card to its owner's hand."
//
// S82OldB489Zendikon, the Zendikon animation and dies trigger Guardian
// Zendikon has.
//
// No simplification.
func init() {
	effects.Register(effects.S82OldB489Zendikon("d6786f80-818c-433d-8978-81814171dd39", "Wind Zendikon", "U", "Elemental", 2, 2, "flying"))
}
