package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blue Scarab — Enchantment — Aura {W} (Ice Age, S82 OLD #5636):
//
//	"Enchant creature
//	 Enchanted creature can't be blocked by blue creatures.
//	 Enchanted creature gets +2/+2 as long as an opponent controls a blue
//	 permanent."
//
// A block rule on the attached creature and a layer 7c pump re-read on every
// recompute, so it starts and stops with the opponent's blue permanents.
//
// No simplification.
func init() {
	effects.Register(effects.ScarabAuraSpecB829("47c12d2f-b458-4848-9a01-0fc37ef3d942", "Blue Scarab", "U", "blue"))
}
