package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The remaining Reality Fracture planeswalker-support duals:
//
//	"This land enters tapped unless you control a planeswalker.
//	 {T}: Add {A} or {B}."
//
// Innovative Commons ({U}/{R}), Konstrari Annex ({R}/{G}) and
// Meticulous Commons ({W}/{B}). One loop, like the rest of the cycle.
// The condition is read once as the land enters (CR 614.12), so a land
// that came in tapped stays tapped when a planeswalker shows up later.
func init() {
	for _, land := range []struct{ oracleID, name, a, b string }{
		{"6812e4d3-0034-4fc8-9ee9-520f76a0ea94", "Innovative Commons", "U", "R"},
		{"31ea4f3f-392a-4319-a95b-eb28790a80e6", "Konstrari Annex", "R", "G"},
		{"7fdd471f-ec95-4aa7-8f8a-ad8946de5ab4", "Meticulous Commons", "W", "B"},
	} {
		Register(Spec{
			OracleID:      land.oracleID,
			Name:          land.name,
			Completeness:  CompletenessFull,
			Replacements:  []game.ReplacementEffect{SelfEntersTappedUnless(youControlAPlaneswalker)},
			ManaAbilities: []ManaAbility{dualManaAbility(land.a, land.b)},
		})
	}
}
