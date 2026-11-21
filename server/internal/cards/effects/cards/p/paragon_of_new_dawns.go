package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Paragon of New Dawns — Creature — Human Soldier {3}{W}, 2/2 (Magic 2015,
// #5556):
//
//	"Other white creatures you control get +1/+1.
//	 {W}, {T}: Another target white creature you control gains vigilance
//	 until end of turn."
//
// The M15 Paragon cycle's shared body (m15B1Paragon).
//
// No simplification.
func init() {
	effects.Register(effects.M15B1Paragon("f064e5e3-4772-4f6d-b8fa-4f36d4fcdb42", "Paragon of New Dawns", "{W}", "W", "white", "vigilance"))
}
