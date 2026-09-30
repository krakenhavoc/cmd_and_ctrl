package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Emerald Medallion — Artifact {2}:
//
//	"Green spells you cast cost {1} less to cast."
//
// The green member of the Tempest Medallion cycle; see
// jet_medallion.go for the shape's mechanics.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d683024d-8fe7-4509-851b-7655fadc1de7",
		Name:         "Emerald Medallion",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Green spells you cast cost {1} less to cast.", YourSpell(), ColoredSpell("G")),
		},
	})
}
