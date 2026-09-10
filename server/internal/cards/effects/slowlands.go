package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// slowlands.go — the Innistrad: Midnight Hunt / Crimson Vow
// "slowland" cycle:
//
//	"This land enters tapped unless you control two or more other
//	 lands."
//	"{T}: Add {X} or {Y}."
//
// Three of the ten, the three this deck plays. The exact inverse of
// a fastland, and deliberately implemented as the same helper with
// the comparison flipped rather than as its own scan — the two
// cycles are one mechanic printed with opposite signs, and a reader
// comparing the files should be able to see that.
func init() {
	for _, t := range []struct {
		oracleID string
		name     string
		a, b     string
	}{
		{"f0ec8681-da50-466b-8cdd-1dc710deccd9", "Deserted Beach", "W", "U"},
		{"c854ecb0-cc60-4c48-a9aa-7f2348a7a8c6", "Shattered Sanctum", "W", "B"},
		{"5f42b67f-87fd-4f98-a0e8-0c8313f4bbc8", "Shipwreck Marsh", "U", "B"},
	} {
		Register(Spec{
			OracleID:      t.oracleID,
			Name:          t.name,
			Replacements:  []game.ReplacementEffect{EntersTappedUnless(otherLandsAtLeast(2))},
			ManaAbilities: []ManaAbility{dualManaAbility(t.a, t.b)},
		})
	}
}
