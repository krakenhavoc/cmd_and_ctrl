package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// tri_lands.go — the Shards of Alara / Khans of Tarkir "tri-lands":
//
//	"This land enters tapped.
//	 {T}: Add {A}, {B}, or {C}."
//
// The six the roadmap's batch 03 (#296) ranks between 361 and 404 —
// every one a three-colour Commander deck's most-played nonbasic
// after the shocks. One table: an unconditional enters-tapped
// self-replacement (the Temple cycle's) and a three-option pipe. The
// pipe narrows to the controller's commander identity at activation
// like every other pipe in the catalog; a three-colour land in a
// three-colour deck is never narrowed.
//
// No simplification.
func init() {
	for _, t := range []struct{ oracleID, name, a, b, c string }{
		{"7190debf-708b-4f41-9714-0d0a5bd5a74e", "Crumbling Necropolis", "U", "B", "R"},
		{"4619de7e-3d6e-4c6b-8e6e-e24db324839d", "Nomad Outpost", "R", "W", "B"},
		{"834b8f71-9a45-42ae-9e99-e749fa6fb45e", "Mystic Monastery", "U", "R", "W"},
		{"f9e7e855-1e3b-42d3-91b0-64ba8b5b8982", "Opulent Palace", "B", "G", "U"},
		{"e4cf6c2f-0f1e-4980-9ef9-e4eabcae42a9", "Frontier Bivouac", "G", "U", "R"},
		{"2ae77795-6a80-498b-bf69-6fd612f601e4", "Seaside Citadel", "G", "W", "U"},
	} {
		Register(Spec{
			OracleID:      t.oracleID,
			Name:          t.name,
			Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
			ManaAbilities: []ManaAbility{b03TriMana(t.a, t.b, t.c)},
		})
	}
}
