package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soured Springs — Land — Desert (Outlaws of Thunder Junction, #4468):
//
//	"This land enters tapped.
//	 When this land enters, it deals 1 damage to target opponent.
//	 {T}: Add {U} or {B}."
//
// The Dimir member of the set's tapped Desert duals (b36DesertDual,
// Abraded Bluffs' shape). Desert is on the printed type line.
//
// No simplification.
func init() {
	effects.Register(effects.B36DesertDual("e579edd7-4f6c-4f22-a72f-0a20d7a698a2", "Soured Springs", "U", "B"))
}
