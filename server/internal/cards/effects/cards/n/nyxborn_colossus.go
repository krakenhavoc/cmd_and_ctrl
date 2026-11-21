package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nyxborn Colossus — Enchantment Creature — Giant {3}{G}{G}{G}, 6/7 (Theros Beyond Death, #5514):
//
//	A vanilla creature: no rules text.
//
// Nothing to implement: the type line and the printed power and
// toughness come from the card data, as for Jasmine Boreal.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4a74cccb-5d75-4362-9c9e-2622096544fe",
		Name:         "Nyxborn Colossus",
		Completeness: effects.CompletenessFull,
	})
}
