package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Coastal Peak — Land — Island Mountain (Marvel Super Heroes Commander,
// #4455): "({T}: Add {U} or {R}.) This land enters tapped. Cycling {2}".
// The blue-red member of the cycling-land cycle, the same shape as
// Canyon Slough.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "0e25faa2-efb0-4ca5-b280-18c38faa860c",
		Name:          "Coastal Peak",
		Completeness:  CompletenessFull,
		Activated:     []ActivatedAbility{Cycling("{2}")},
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility("U", "R")},
	})
}
