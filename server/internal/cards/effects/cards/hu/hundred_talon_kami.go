package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hundred-Talon Kami — Creature — Spirit {4}{W}, 2/3 (Champions of
// Kamigawa, #5603):
//
//	"Flying
//	 Soulshift 4"
//
// No simplification.
func init() {
	effects.ChkB1Soulshifter("100b4440-c588-423e-93d2-23c9b595614d", "Hundred-Talon Kami", 4, "flying")
}
