package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Behemoth's Herald — Creature — Elf Shaman {G}, 1/1 (Shards of Alara, S82
// OLD batch 509, #5588):
//
//	"{2}{G}, {T}, Sacrifice a red creature, a green creature, and a white
//	 creature: Search your library for a card named Godsire, put it onto
//	 the battlefield, then shuffle."
func init() {
	effects.Register(effects.S82OldB509Herald("967c509f-dbde-44ee-93fa-a01376f34b99", "Behemoth's Herald", "{2}{G}", "Godsire",
		effects.S82OldB509ColoredCreature("a red creature", "R"),
		effects.S82OldB509ColoredCreature("a green creature", "G"),
		effects.S82OldB509ColoredCreature("a white creature", "W")))
}
