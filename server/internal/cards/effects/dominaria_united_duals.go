package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// dominaria_united_duals.go — the Dominaria United typed taplands:
//
//	"({T}: Add {B} or {G}.)
//	 This land enters tapped."
//
// Land — Swamp Forest, so the basic land TYPES carry the mana
// ability as reminder text and — as with the original duals — the
// engine will not synthesise it for a nonbasic, which is why the
// cycle needs a table at all. The tapped entry is the real CR 614
// self-replacement. Roadmap batch 11 (#304) ranks Haunted Mire and
// batch 12 (#305) Tangled Islet, batch 16 (#309) Sacred Peaks, Sunlit
// Marsh and Wooded Ridgeline; the other five belong here when their
// batches reach them. Foundations reprinted the shape as a new
// cycle with new names — Radiant Grove (batch 13, #306) is that
// cycle's Forest Plains, identical text, so it is a row here rather
// than a second table.
//
// No simplification.
func init() {
	for _, land := range []struct{ oracleID, name, a, b string }{
		{"b0b58a03-462c-4964-97c7-42bc777ec23e", "Haunted Mire", "B", "G"},
		// Roadmap batch 12 (#305): Land — Forest Island.
		{"4b1f68a2-b606-4c64-bd44-a9714808316d", "Tangled Islet", "G", "U"},
		// Roadmap batch 13 (#306): Foundations' Land — Forest Plains.
		{"32c91719-f3dd-4cc7-9e32-7d5ccf18f07c", "Radiant Grove", "G", "W"},
		// Roadmap batch 16 (#309): Land — Mountain Plains, Land —
		// Plains Swamp, Land — Mountain Forest.
		{"fb69bc57-f05a-41c2-9b7b-9a9761ef0cd3", "Sacred Peaks", "R", "W"},
		{"a5e5a259-5fa7-4b01-93cb-a2b4aaf80927", "Sunlit Marsh", "W", "B"},
		{"c2ca3e20-23ca-4d2a-88a1-5e98ff884abb", "Wooded Ridgeline", "R", "G"},
	} {
		Register(Spec{
			OracleID:      land.oracleID,
			Name:          land.name,
			Completeness:  CompletenessFull,
			Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
			ManaAbilities: []ManaAbility{dualManaAbility(land.a, land.b)},
		})
	}
}
