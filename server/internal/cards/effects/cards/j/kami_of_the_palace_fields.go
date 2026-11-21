package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kami of the Palace Fields — Creature — Spirit {5}{W}, 3/2 (Champions of
// Kamigawa, #5603):
//
//	"Flying, first strike
//	 Soulshift 5"
//
// No simplification.
func init() {
	effects.ChkB1Soulshifter("e8d44867-9519-41f2-aa0f-4079444aeef9", "Kami of the Palace Fields", 5, "flying", "first strike")
}
