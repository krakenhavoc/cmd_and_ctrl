package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Furious Resistance — Instant {R} (Gatecrash, S82 OLD #5565):
//
//	"Target blocking creature gets +3/+0 and gains first strike until end of
//	 turn."
//
// Captain's Defense's blocking-creature target (CR 506.4) with Leap of Flame's
// pump-and-grant body. A target that stopped blocking is illegal on resolution
// and the spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e69a6fee-c4c0-4b31-8899-6e0935a6cded",
		Name:         "Furious Resistance",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target blocking creature", effects.BlockingCreature()),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 0, "first strike")),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.PumpAndGrantFirstTarget(3, 0, "Furious Resistance — +3/+0 and first strike", "first strike")(ctx.Game, item)
		},
	})
}
