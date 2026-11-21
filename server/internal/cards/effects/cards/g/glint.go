package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glint — Instant {1}{U} (Dragons of Tarkir,
// #5551):
//
//	"Target creature you control gets +0/+3 and gains hexproof until end of turn. (It can't be the target of spells or abilities your opponents control.)"
//
// Dive Down's body: +0/+3 and hexproof until end of turn on the target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2fc07680-3ffb-4507-a1e1-b9580f5ce878",
		Name:         "Glint",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 3, "hexproof")),
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.PumpAndGrantFirstTarget(0, 3, "Glint — +0/+3 and hexproof", "hexproof")(ctx.Game, item)
		},
	})
}
