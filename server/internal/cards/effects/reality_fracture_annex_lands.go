package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// reality_fracture_annex_lands.go — the four Reality Fracture lands that
// print
//
//	"This land enters tapped unless you control a planeswalker.
//	 {T}: Add {A} or {B}."
//
// Stingerquill Annex ({B}/{R}), Theorix Annex ({U}/{B}), Transformative
// Commons ({G}/{U}) and Vigorbloom Annex ({G}/{W}). One loop over a table,
// as the battle lands do, since four near-identical files are four places
// to fix one mistake. The condition is a self-replacement
// (SelfEntersTappedUnless), so the land never enters untapped and is then
// tapped, and a planeswalker token (a Jace token) counts as a planeswalker.
//
// No simplification.
func init() {
	for _, land := range []struct{ oracleID, name, a, b string }{
		{"26e20ae5-4059-4d63-9db5-6e12421d1aba", "Stingerquill Annex", "B", "R"},
		{"00ad5200-5179-4d1b-8bb3-a09e1b31be58", "Theorix Annex", "U", "B"},
		{"8ec56e2a-380f-4de7-9fb7-52de803008fc", "Transformative Commons", "G", "U"},
		{"fcb35d42-bcd2-427c-9640-ff3602b79912", "Vigorbloom Annex", "G", "W"},
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
