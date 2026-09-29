package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "f1ef5698-14bd-42bf-aae5-2870699e4186",
		Name:         "Sapphire Medallion",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Blue spells you cast cost {1} less to cast.", YourSpell(), ColoredSpell("U")),
		},
	})
}
