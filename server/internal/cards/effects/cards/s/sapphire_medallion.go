package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sapphire Medallion — Artifact {2}:
//
//	"Blue spells you cast cost {1} less to cast."
//
// The blue member of the Tempest Medallion cycle; see jet_medallion.go
// for the shape. The reduction spends generic mana only (CR 601.2f),
// so a {U} spell costs {U} under it and a {1}{U} spell costs {U}. A
// multicolour spell with blue in it is a blue spell (CR 105.2) and is
// discounted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f1ef5698-14bd-42bf-aae5-2870699e4186",
		Name:         "Sapphire Medallion",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Blue spells you cast cost {1} less to cast.", effects.YourSpell(), effects.ColoredSpell("U")),
		},
	})
}
