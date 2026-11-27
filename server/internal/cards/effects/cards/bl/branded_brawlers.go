package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Branded Brawlers — Creature — Human Soldier {R}, 2/2 (Prophecy, S82 OLD b671,
// #5617):
//
//	"This creature can't attack if defending player controls an untapped land.
//	 This creature can't block if you control an untapped land."
//
// The shared Brawlers body (OldB671BrawlerSpec).
//
// No simplification.
func init() {
	effects.Register(effects.OldB671BrawlerSpec("e975a8b3-f817-42f3-84da-b03ebf95cf25", "Branded Brawlers"))
}
