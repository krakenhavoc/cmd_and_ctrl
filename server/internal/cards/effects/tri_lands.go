package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// tri_lands.go — the Shards of Alara tri-lands:
//
//	"This land enters tapped.
//	 {T}: Add {W}, {U}, or {B}."
//
// Raffine's Tower's shape without the cycling: a real CR 614
// self-replacement for the tapped entry and one three-colour pipe,
// so the activation is one picker rather than three menu entries.
// Roadmap batch 02 (#295) ranks Arcane Sanctum and Jungle Shrine;
// the other three belong in this table when their batches arrive.
//
// No simplification.
func init() {
	for _, land := range []struct{ oracleID, name, pipe, label string }{
		{"7d7cf15c-06b9-4062-a1eb-32614c458a3b", "Arcane Sanctum", "{W|U|B}", "Add {W}, {U}, or {B}"},
		{"2e69537c-c898-4e13-a72d-ce3957a90304", "Jungle Shrine", "{R|G|W}", "Add {R}, {G}, or {W}"},
	} {
		Register(Spec{
			OracleID:     land.oracleID,
			Name:         land.name,
			Completeness: CompletenessFull,
			Replacements: []game.ReplacementEffect{SelfEntersTapped()},
			ManaAbilities: []ManaAbility{{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: land.pipe,
				Label:    land.label,
			}},
		})
	}
}
