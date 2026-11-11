package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sage of the Beyond — Creature — Spirit Giant {5}{U}{U}, 5/5 (Outlaws of
// Thunder Junction Commander):
//
//	"Flying
//	 Spells you cast from anywhere other than your hand cost {2} less to
//	 cast.
//	 Foretell {4}{U} (...)"
//
// The reduction is the shared cmdr2S40CastFromOutsideHandCostsLess, read
// from the zone the spell is cast from (CR 601.2f), so a commander from the
// command zone, a flashback card and a foretold card all cost {2} less,
// and a hand cast does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fc74f6a6-9064-4e1b-8dc1-446192bbc448",
		Name:            "Sage of the Beyond",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SpecialActions:  []game.SpecialAction{Foretell("{4}{U}")},
		CostModifiers: []game.CostModifier{
			cmdr2S40CastFromOutsideHandCostsLess(2, "Spells you cast from anywhere other than your hand cost {2} less to cast."),
		},
	})
}
