package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// tri_lands.go — the Shards of Alara / Khans of Tarkir "tri-lands"
// the roadmap's batch 02 (#295) ranks in the top 360:
//
//	"This land enters tapped."
//	"{T}: Add {W}, {U}, or {B}."
//
// Plain taplands with a three-colour ability. Unlike the Triomes
// (Raffine's Tower, Ketria Triome) these carry NO land types at all,
// so there is no cycling, nothing fetches them, and the entire card
// is the two lines above.
//
// SelfEntersTapped is a real CR 614 self-replacement rather than an
// OnETB tap: the land is never untapped on the battlefield, so
// nothing watching for a tap event or for an untapped permanent
// entering sees the wrong thing. (Azorius Chancery's file still
// carries a note saying a catalog self-replacement cannot fire on
// its own source's entry — that was true when it was written and
// stopped being true with the Temple cycle, which added the
// entering card's own replacements as a third gathering block.)
//
// One pipe ability rather than three menu entries, matching
// Raffine's Tower. Commander-identity narrowing is left on: a
// three-colour land is only legal in a deck whose identity already
// covers all three, so the intersection is a no-op here.
//
// No simplifications.
func init() {
	for _, land := range []struct{ oracleID, name, a, b, c string }{
		{"7d7cf15c-06b9-4062-a1eb-32614c458a3b", "Arcane Sanctum", "W", "U", "B"},
		{"2e69537c-c898-4e13-a72d-ce3957a90304", "Jungle Shrine", "R", "G", "W"},
	} {
		Register(Spec{
			OracleID:     land.oracleID,
			Name:         land.name,
			Replacements: []game.ReplacementEffect{SelfEntersTapped()},
			ManaAbilities: []ManaAbility{{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{" + land.a + "|" + land.b + "|" + land.c + "}",
				Label:    "Add {" + land.a + "}, {" + land.b + "}, or {" + land.c + "}",
			}},
		})
	}
}
