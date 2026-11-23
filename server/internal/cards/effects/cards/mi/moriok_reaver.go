package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moriok Reaver — Creature — Human Warrior {2}{B}, 3/2 (Scars of Mirrodin, #5577)
//
// No rules text: a vanilla body. Registered so the catalogue lists it
// as reviewed; its cost, type line and power and toughness are printed
// card data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9bd62fa0-b99c-4b64-a6f6-608cf01f8ef3",
		Name:         "Moriok Reaver",
		Completeness: effects.CompletenessFull,
	})
}
