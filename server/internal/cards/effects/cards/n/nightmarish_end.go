package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nightmarish End — Instant {2}{B} (Journey into Nyx, S82 OLD, #5558):
//
//	"Target creature gets -X/-X until end of turn, where X is the number of
//	 cards in your hand."
//
// X is counted as the spell resolves (CR 608.2h); the spell itself is on the
// stack and so is not in the hand. A target that has left does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "70f5e932-0dc5-4c34-b7be-16cd80ec10b7",
		Name:         "Nightmarish End",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, 0, 0)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
			if p == nil || p.Hand == nil {
				return nil
			}
			x := p.Hand.Size()
			return effects.VsPumpTheTarget(ctx, -x, -x, "Nightmarish End — -X/-X")
		},
	})
}
