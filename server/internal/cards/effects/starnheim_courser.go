package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Starnheim Courser — {2}{W} Creature — Pegasus 2/2 (Foundations
// Jumpstart):
//
//	"Flying
//	 Artifact and enchantment spells you cast cost {1} less to cast."
//
// One cost reduction per spell, whichever of the two types it has; an
// artifact enchantment is discounted once (CR 601.2f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "01317d6c-8c11-40e1-8f7b-d36050973ada",
		Name:            "Starnheim Courser",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Artifact and enchantment spells you cast cost {1} less to cast.",
				YourSpell(), func(q game.CostQuery) bool {
					return q.Card.HasCardType("artifact") || q.Card.HasCardType("enchantment")
				}),
		},
	})
}
