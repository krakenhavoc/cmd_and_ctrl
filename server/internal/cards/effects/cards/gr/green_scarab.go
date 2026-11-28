package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Green Scarab — Enchantment — Aura {W} (Ice Age, S82 OLD #5636):
//
//	"Enchant creature
//	 Enchanted creature can't be blocked by green creatures.
//	 Enchanted creature gets +2/+2 as long as an opponent controls a green
//	 permanent."
//
// A block rule on the attached creature and a layer 7c pump re-read on every
// recompute.
//
// No simplification.
func init() {
	effects.Register(effects.ScarabAuraSpecB830("c5ca2df3-6fd9-478d-9bc6-48695a4600fb", "Green Scarab", "G", "green"))
}
