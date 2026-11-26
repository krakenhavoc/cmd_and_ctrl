package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Corrupted Zendikon — Enchantment — Aura {1}{B} (Worldwake, S82 OLD batch 489, #5582):
//
//	"Enchant land
//	 Enchanted land is a 3/3 black Ooze creature. It's still a land.
//	 When enchanted land dies, return that card to its owner's hand."
//
// S82OldB489Zendikon, the Zendikon animation and dies trigger Guardian
// Zendikon has.
//
// No simplification.
func init() {
	effects.Register(effects.S82OldB489Zendikon("3902048e-b7bf-4676-be92-e590bc76c729", "Corrupted Zendikon", "B", "Ooze", 3, 3, ""))
}
