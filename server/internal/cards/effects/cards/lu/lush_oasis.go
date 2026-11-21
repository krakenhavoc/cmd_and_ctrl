package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lush Oasis — Land — Desert (Outlaws of Thunder Junction, #4468):
//
//	"This land enters tapped.
//	 When this land enters, it deals 1 damage to target opponent.
//	 {T}: Add {G} or {U}."
//
// The Simic member of Outlaws of Thunder Junction's tapped Desert duals,
// Eroded Canyon's shape exactly, built by b36DesertDual.
//
// No simplification.
func init() {
	effects.Register(effects.B36DesertDual("b6a965eb-cffb-41c1-925a-7cf3e8e2f248", "Lush Oasis", "G", "U"))
}
