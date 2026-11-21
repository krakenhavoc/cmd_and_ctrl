package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ruby Medallion — Artifact {2} (EDHREC rank 293):
//
//	"Red spells you cast cost {1} less to cast."
//
// See jet_medallion.go for the shape's mechanics — this is the same
// body for red spells.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0c367958-729e-416d-988b-098e90e7a1fd",
		Name:         "Ruby Medallion",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Red spells you cast cost {1} less to cast.", effects.YourSpell(), effects.ColoredSpell("R")),
		},
	})
}
