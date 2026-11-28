package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// White Scarab — Enchantment — Aura {W} (Ice Age, S82 OLD #5636):
//
//	"Enchant creature
//	 Enchanted creature can't be blocked by white creatures.
//	 Enchanted creature gets +2/+2 as long as an opponent controls a white
//	 permanent."
//
// A block rule on the attached creature and a layer 7c pump re-read on every
// recompute.
//
// No simplification.
func init() {
	effects.Register(effects.ScarabAuraSpecB830("ec9c3906-8628-48c6-96a7-8c6748cba816", "White Scarab", "W", "white"))
}
