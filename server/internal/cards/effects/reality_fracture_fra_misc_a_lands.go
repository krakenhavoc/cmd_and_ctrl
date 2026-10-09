package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// reality_fracture_fra_misc_a_lands.go — the three Reality Fracture
// "catch-up" typed duals:
//
//	"({T}: Add {A} or {B}.)
//	 This land enters tapped unless your opponents control eight or
//	 more lands."
//
// Turbulent Crater ({B}/{R}, Swamp Mountain), Turbulent Shore ({W}/{U},
// Plains Island) and Turbulent Wetlands ({U}/{B}, Island Swamp). Same
// body as Turbulent Fen and Turbulent Springs (turbulent_fen.go): a
// CR 614 self-replacement on the shared b40CatchUpDualCondition and a
// pipe-syntax dual, as one loop over a table so the cycle's threshold
// is written once.
//
// No simplification.
func init() {
	for _, land := range []struct{ oracleID, name, a, b string }{
		{"e86a0b64-fe4a-4ea8-99e8-35860dbac765", "Turbulent Crater", "B", "R"},
		{"c13c5f07-dac5-47d0-a088-e24b9fbec3c5", "Turbulent Shore", "W", "U"},
		{"b51c8659-ec4f-4213-beb1-49a7acd7366c", "Turbulent Wetlands", "U", "B"},
	} {
		Register(Spec{
			OracleID:      land.oracleID,
			Name:          land.name,
			Completeness:  CompletenessFull,
			Replacements:  []game.ReplacementEffect{SelfEntersTappedUnless(b40CatchUpDualCondition())},
			ManaAbilities: []ManaAbility{dualManaAbility(land.a, land.b)},
		})
	}
}
