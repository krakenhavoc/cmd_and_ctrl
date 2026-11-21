package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ni'Var Science Academy — Land (Star Trek, #4973):
//
//	"This land enters tapped unless you control two or more other lands.
//	 When this land enters untapped, choose one —
//	 • Target creature gains flying until end of turn.
//	 • Surveil 1.
//	 {T}: Add {U}."
//
// The blue member of Star Trek's surveil-land cycle (trekB29SurveilLand);
// its bullet is a keyword grant with no pump.
//
// No simplification.
func init() {
	effects.TrekB29SurveilLand("45bf09bb-4141-4166-bdae-6dac39e7f316", "Ni'Var Science Academy", "U",
		effects.TrekB29TargetCreatureUntilEOT("Target creature gains flying until end of turn.", 0, 0, "flying"))
}
