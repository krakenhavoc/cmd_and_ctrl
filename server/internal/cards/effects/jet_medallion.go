package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jet Medallion — Artifact {2} (EDHREC rank 278):
//
//	"Black spells you cast cost {1} less to cast."
//
// A board-wide cost modifier over the controller's own black spells —
// CostsLess with YourSpell() and ColoredSpell("B"), the same
// vocabulary the roadmap's cost-modification group (#93) shipped for
// the whole Medallion cycle's siblings.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bccb06f1-8839-4104-9ffa-c79a817f8378",
		Name:         "Jet Medallion",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Black spells you cast cost {1} less to cast.", YourSpell(), ColoredSpell("B")),
		},
	})
}
