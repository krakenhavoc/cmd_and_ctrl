package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Reality Fracture planeswalker-support duals:
//
//	"This land enters tapped unless you control a planeswalker.
//	 {T}: Add {A} or {B}."
//
// Dedicated Commons ({R}/{W}), Fatehold Annex ({W}/{U}) and Formidable
// Commons ({B}/{G}). One loop rather than three files, the way the
// battle lands are written: three near-identical files are three places
// to fix one mistake. The condition is checked once, as the land
// enters, and never again (CR 614.12), so a land that came in tapped
// stays tapped when a planeswalker shows up later.
func init() {
	for _, land := range []struct{ oracleID, name, a, b string }{
		{"d4e0749a-e179-4c19-8536-bf8f1a198549", "Dedicated Commons", "R", "W"},
		{"d35876a3-e891-43d3-a50c-69e8469497e0", "Fatehold Annex", "W", "U"},
		{"3e51c060-1e86-4a0e-8fee-8b7273161468", "Formidable Commons", "B", "G"},
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
