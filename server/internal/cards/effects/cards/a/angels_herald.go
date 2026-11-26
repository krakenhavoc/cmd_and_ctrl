package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Angel's Herald — Creature — Human Cleric {W}, 1/1 (Shards of Alara, S82
// OLD batch 509, #5588):
//
//	"{2}{W}, {T}, Sacrifice a green creature, a white creature, and a blue
//	 creature: Search your library for a card named Empyrial Archangel, put
//	 it onto the battlefield, then shuffle."
func init() {
	effects.Register(effects.S82OldB509Herald("9e4db334-4698-4986-820f-1daf1bffda83", "Angel's Herald", "{2}{W}", "Empyrial Archangel",
		effects.S82OldB509ColoredCreature("a green creature", "G"),
		effects.S82OldB509ColoredCreature("a white creature", "W"),
		effects.S82OldB509ColoredCreature("a blue creature", "U")))
}
