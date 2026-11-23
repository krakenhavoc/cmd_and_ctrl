package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nyxborn Courser — Enchantment Creature — Centaur Scout {1}{W}{W}, 2/4 (Theros Beyond Death, #5514):
//
//	A vanilla creature: no rules text.
//
// Nothing to implement: the type line and the printed power and
// toughness come from the card data, as for Jasmine Boreal.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a9b345ea-3e34-48cf-8ab8-887c2b67df9c",
		Name:         "Nyxborn Courser",
		Completeness: effects.CompletenessFull,
	})
}
