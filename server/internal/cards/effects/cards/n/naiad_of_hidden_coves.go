package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Naiad of Hidden Coves — {2}{U} Enchantment Creature — Nymph 2/3:
//
//	"During turns other than yours, spells you cast cost {1} less to cast."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "38b4798d-47cc-4ffd-8b0b-e90b5da2ee2d",
		Name:         "Naiad of Hidden Coves",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "During turns other than yours, spells you cast cost {1} less to cast.",
				effects.YourSpell(), effects.CastDuringAnotherPlayersTurn()),
		},
	})
}
