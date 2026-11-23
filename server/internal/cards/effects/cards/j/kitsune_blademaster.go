package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kitsune Blademaster — Creature — Fox Samurai {2}{W}, 2/2 (Champions of
// Kamigawa, #5603):
//
//	"First strike
//	 Bushido 1"
//
// No simplification.
func init() {
	effects.ChkB1Bushido("2db73fff-1a89-4006-945d-a244ff2f0bef", "Kitsune Blademaster", 1, "first strike")
}
