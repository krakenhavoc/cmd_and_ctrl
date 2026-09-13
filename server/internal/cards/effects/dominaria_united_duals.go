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
// batch 12 (#305) Tangled Islet; the other eight belong here when
// their batches reach them.
//
// No simplification.
func init() {
	for _, land := range []struct{ oracleID, name, a, b string }{
		{"b0b58a03-462c-4964-97c7-42bc777ec23e", "Haunted Mire", "B", "G"},
		// Roadmap batch 12 (#305): Land — Forest Island.
		{"4b1f68a2-b606-4c64-bd44-a9714808316d", "Tangled Islet", "G", "U"},
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
