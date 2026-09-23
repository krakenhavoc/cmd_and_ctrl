package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ruby Medallion — Artifact {2} (EDHREC rank 293):
//
//	"Red spells you cast cost {1} less to cast."
//
// See jet_medallion.go for the shape's mechanics — this is the same
// body for red spells.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0c367958-729e-416d-988b-098e90e7a1fd",
		Name:         "Ruby Medallion",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Red spells you cast cost {1} less to cast.", YourSpell(), ColoredSpell("R")),
		},
	})
}
