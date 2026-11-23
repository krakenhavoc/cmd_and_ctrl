package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gift of Strength — Instant {1}{G} (Ravnica: Clue Edition, #5929):
//
//	"Target creature gets +3/+3 and gains reach until end of turn."
//
// Dive Down's body: the boost and the keyword land on the still-legal
// target (pumpAndGrantFirstTarget).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fe971240-268a-47d4-a14e-3d6352656091",
		Name:         "Gift of Strength",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3, "reach")),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.PumpAndGrantFirstTarget(3, 3, "Gift of Strength — +3/+3 and reach", "reach")(ctx.Game, item)
		},
	})
}
