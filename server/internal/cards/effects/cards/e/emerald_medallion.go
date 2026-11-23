package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Emerald Medallion — Artifact {2}:
//
//	"Green spells you cast cost {1} less to cast."
//
// The green member of the Tempest Medallion cycle; see
// jet_medallion.go for the shape's mechanics.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d683024d-8fe7-4509-851b-7655fadc1de7",
		Name:         "Emerald Medallion",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Green spells you cast cost {1} less to cast.", effects.YourSpell(), effects.ColoredSpell("G")),
		},
	})
}
