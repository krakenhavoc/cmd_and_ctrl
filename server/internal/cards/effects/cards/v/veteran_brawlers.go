package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Veteran Brawlers — Creature — Human Soldier {1}{R}, 4/4 (Prophecy, S82 OLD b671,
// #5617):
//
//	"This creature can't attack if defending player controls an untapped land.
//	 This creature can't block if you control an untapped land."
//
// The shared Brawlers body (OldB671BrawlerSpec).
//
// No simplification.
func init() {
	effects.Register(effects.OldB671BrawlerSpec("77a963a3-473a-4c34-bc46-8ef016a92c0c", "Veteran Brawlers"))
}
