package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arcane Epiphany — Instant {3}{U}{U} (Foundations, #4465):
//
//	"This spell costs {1} less to cast if you control a Wizard.
//	 Draw three cards."
//
// Wizard's Retort's conditional self cost modifier (#746), then a draw
// of three.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fe1556cc-32d0-4e51-bb15-f3fce669a5e1",
		Name:         "Arcane Epiphany",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 3},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(1, "This spell costs {1} less to cast if you control a Wizard.", effects.YouControlA(effects.HasSubtype("Wizard"))),
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: item.Controller, N: 3}.Apply(ctx)
		},
	})
}
