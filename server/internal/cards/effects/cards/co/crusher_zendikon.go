package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crusher Zendikon — Enchantment — Aura {2}{R} (Worldwake, S82 OLD batch 489, #5582):
//
//	"Enchant land
//	 Enchanted land is a 4/2 red Beast creature with trample. It's still a land.
//	 When enchanted land dies, return that card to its owner's hand."
//
// S82OldB489Zendikon, the Zendikon animation and dies trigger Guardian
// Zendikon has.
//
// No simplification.
func init() {
	effects.Register(effects.S82OldB489Zendikon("68a1a67e-9d40-47dd-b282-6c1b15465f62", "Crusher Zendikon", "R", "Beast", 4, 2, "trample"))
}
