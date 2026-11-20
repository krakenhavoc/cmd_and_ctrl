package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tranquil Expanse — Land (Oath of the Gatewatch, #5545):
//
//	"This land enters tapped.
//	 {T}: Add {G} or {W}."
//
// The Guildgates' shape: it enters tapped (a replacement, CR 614.1c) and
// taps for either colour.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "a5478263-47c9-447f-9c7a-c77ce0752947",
		Name:          "Tranquil Expanse",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility("G", "W")},
	})
}
