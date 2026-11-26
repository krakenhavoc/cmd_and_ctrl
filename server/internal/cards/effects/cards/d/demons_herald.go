package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Demon's Herald — Creature — Human Wizard {B}, 1/1 (Shards of Alara, S82
// OLD batch 509, #5588):
//
//	"{2}{B}, {T}, Sacrifice a blue creature, a black creature, and a red
//	 creature: Search your library for a card named Prince of Thralls, put
//	 it onto the battlefield, then shuffle."
func init() {
	effects.Register(effects.S82OldB509Herald("9f4e4d6e-bb1d-41d6-9184-3cac45197285", "Demon's Herald", "{2}{B}", "Prince of Thralls",
		effects.S82OldB509ColoredCreature("a blue creature", "U"),
		effects.S82OldB509ColoredCreature("a black creature", "B"),
		effects.S82OldB509ColoredCreature("a red creature", "R")))
}
