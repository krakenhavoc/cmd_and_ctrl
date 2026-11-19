package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Naiad of Hidden Coves — {2}{U} Enchantment Creature — Nymph 2/3:
//
//	"During turns other than yours, spells you cast cost {1} less to cast."
func init() {
	Register(Spec{
		OracleID:     "38b4798d-47cc-4ffd-8b0b-e90b5da2ee2d",
		Name:         "Naiad of Hidden Coves",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "During turns other than yours, spells you cast cost {1} less to cast.",
				YourSpell(), CastDuringAnotherPlayersTurn()),
		},
	})
}
