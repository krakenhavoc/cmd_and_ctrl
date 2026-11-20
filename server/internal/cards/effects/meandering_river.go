package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Meandering River — Land (Oath of the Gatewatch, #5545):
//
//	"This land enters tapped.
//	 {T}: Add {W} or {U}."
//
// The Guildgates' shape: it enters tapped (a replacement, CR 614.1c) and
// taps for either colour.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "2d9663be-c466-4191-85e2-a69ce0965432",
		Name:          "Meandering River",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility("W", "U")},
	})
}
