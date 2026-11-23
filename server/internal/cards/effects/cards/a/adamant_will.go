package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Adamant Will — Instant {1}{W} (Foundations, #4465):
//
//	"Target creature gets +2/+2 and gains indestructible until end of
//	 turn."
//
// Hamato Guardian Stance's shape without the scry: the pump and the
// keyword land on the still-legal target (pumpAndGrantFirstTarget).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "467e22c3-6107-40ff-afc7-5960710c970b",
		Name:         "Adamant Will",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2, "indestructible")),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.PumpAndGrantFirstTarget(2, 2, "Adamant Will — +2/+2 and indestructible", "indestructible")(ctx.Game, item)
		},
	})
}
