package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pearl Medallion — Artifact {2}:
//
//	"White spells you cast cost {1} less to cast."
//
// The white member of the Tempest Medallion cycle; see
// jet_medallion.go for the shape. The reduction spends generic mana
// only (CR 601.2f), so a {W} spell costs {W} under it and a {1}{W}
// spell costs {W}. A multicolour spell with white in it is a white
// spell (CR 105.2) and is discounted.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "19015380-1332-4960-8cc6-0732009525a2",
		Name:         "Pearl Medallion",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "White spells you cast cost {1} less to cast.", YourSpell(), ColoredSpell("W")),
		},
	})
}
