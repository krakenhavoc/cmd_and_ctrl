package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alpha Tyrranax — Creature — Dinosaur Beast {4}{G}{G}, 6/5 (Scars of Mirrodin, #5577)
//
// No rules text: a vanilla body. Registered so the catalogue lists it
// as reviewed; its cost, type line and power and toughness are printed
// card data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8bed1349-4d0a-4544-ba66-c38e72cd2863",
		Name:         "Alpha Tyrranax",
		Completeness: effects.CompletenessFull,
	})
}
