package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// White Shield Crusader — Creature — Human Knight {W}{W}, 2/1 (Coldsnap, S82 OLD b560):
//
//	"Protection from black
//	 {W}: This creature gains flying until end of turn.
//	 {W}{W}: This creature gets +1/+0 until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.OldB560Crusader("96ef0bd9-4abb-49af-8744-0b2ad595a393", "White Shield Crusader", "W", "black"))
}
