package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bellowing Tanglewurm — Creature — Wurm {3}{G}{G}, 4/4 (Scars of
// Mirrodin, #5577):
//
//	"Intimidate
//	 Other green creatures you control have intimidate."
//
// Intimidate is the engine's block rule (CR 702.13b, block_legality.go).
// The grant is Glen Elendra Liege's "other creatures you control of a
// colour" as a layer 6 keyword grant: a creature that turns green later
// has it at once, and one that stops being green loses it.
//
// No simplification.
func init() {
	grant := KeywordGrant(g2OtherCreaturesYouControlOfColor("G"), "intimidate")
	grant.Label = "Other green creatures you control have intimidate."
	Register(Spec{
		OracleID:        "a1091651-4e3e-4840-87e2-b80084770137",
		Name:            "Bellowing Tanglewurm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"intimidate"},
		Static:          []game.StaticAbility{grant},
	})
}
