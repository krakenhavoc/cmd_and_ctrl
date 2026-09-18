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
// Marsh and Wooded Ridgeline, batch 20 (#313) Contaminated Aquifer,
// batch 21 (#383) Idyllic Beachfront and Geothermal Bog, batch 24
// (#386) Molten Tributary; the last one belongs here when its batch
// reaches it. Foundations
// reprinted the shape as a new cycle with new names — Radiant Grove
// (batch 13, #306) is that cycle's Forest Plains, identical text, so
// it is a row here rather than a second table.
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
		// Roadmap batch 20 (#313): Land — Island Swamp.
		{"c27b771d-b5ec-459a-a101-f078cb8d0184", "Contaminated Aquifer", "U", "B"},
		// Roadmap batch 21 (#383): Land — Plains Island, Land — Swamp
		// Mountain.
		{"0aea7e5d-40d7-46c7-a8a6-479cb8061e49", "Idyllic Beachfront", "W", "U"},
		{"e3b67368-1dd6-419b-a95d-7131b1dba23f", "Geothermal Bog", "B", "R"},
		// Roadmap batch 24 (#386): Land — Island Mountain.
		{"58c592ed-20fc-481b-909b-2315567e5f20", "Molten Tributary", "U", "R"},
		// Roadmap batch 25 (#387): Kaldheim's Snow Land — Forest
		// Island. Same text as the rest of the table (Radiant Grove's
		// precedent for a second cycle with identical text); the snow
		// supertype is printed type-line data the engine carries
		// unchanged, so nothing here needs to know about it.
		{"983739cd-0b36-40d9-9a03-7b6aa7ffd0df", "Rimewood Falls", "G", "U"},
		// Roadmap batch 30 (#393): Kaldheim's Snow Land — Swamp Forest,
		// the same cycle as Rimewood Falls.
		{"bf5482b6-dd3e-4fb7-bc62-29e23b417a5f", "Woodland Chasm", "B", "G"},
		// Roadmap batch 32 (#395): Kaldheim's Snow Land — Mountain
		// Forest, the same cycle as Rimewood Falls.
		{"35137378-6754-4bb1-a38e-5940890ccab1", "Highland Forest", "R", "G"},
		// Roadmap batch 42 (#449): Kaldheim's Snow Land — Mountain
		// Plains, the same cycle as Rimewood Falls.
		{"8c281ebe-d9a1-48af-b58b-19c55aa4625b", "Alpine Meadow", "R", "W"},
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
