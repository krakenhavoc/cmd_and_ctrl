package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// slowlands.go — the Innistrad: Midnight Hunt / Crimson Vow
// "slowland" cycle:
//
//	"This land enters tapped unless you control two or more other
//	 lands."
//	"{T}: Add {X} or {Y}."
//
// Six of the ten: the three the Hashaton deck plays, plus the three
// the roadmap's batch 01 (#294) ranks in the top 230 — Dreamroot
// Cascade, Stormcarved Coast and Rockfall Vale. The exact inverse of
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
		// Hashaton deck (#258).
		{"f0ec8681-da50-466b-8cdd-1dc710deccd9", "Deserted Beach", "W", "U"},
		{"c854ecb0-cc60-4c48-a9aa-7f2348a7a8c6", "Shattered Sanctum", "W", "B"},
		{"5f42b67f-87fd-4f98-a0e8-0c8313f4bbc8", "Shipwreck Marsh", "U", "B"},
		// Roadmap batch 01 (#294).
		{"dd8538e6-cd5f-4a88-aff5-eb5e76ce8ddb", "Dreamroot Cascade", "G", "U"},
		{"4722105b-0085-4bb8-bca1-9de0d3eb5600", "Stormcarved Coast", "U", "R"},
		{"185c70c1-8403-4ae5-b45d-3679d4ee092a", "Rockfall Vale", "R", "G"},
	} {
		Register(Spec{
			OracleID:      t.oracleID,
			Name:          t.name,
			Replacements:  []game.ReplacementEffect{EntersTappedUnless(otherLandsAtLeast(2))},
			ManaAbilities: []ManaAbility{dualManaAbility(t.a, t.b)},
		})
	}
}
