package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gibbering Kami — Creature — Spirit {3}{B}, 2/2 (Champions of Kamigawa,
// #5603):
//
//	"Flying
//	 Soulshift 3"
//
// No simplification.
func init() {
	effects.ChkB1Soulshifter("8a5610bd-5f04-43f3-9e2e-d6c728ee899b", "Gibbering Kami", 3, "flying")
}
