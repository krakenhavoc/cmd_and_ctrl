package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aerial Maneuver — Instant {1}{W} (Gatecrash, S82 OLD #5565):
//
//	"Target creature gets +1/+1 and gains flying and first strike until end of
//	 turn."
//
// Leap of Flame's pump-and-grant body (PumpAndGrantFirstTarget) at +1/+1: both
// halves go to the target as the spell resolves, and a target gone by then
// makes it do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d4ccc319-db23-4ef4-8129-317426f3a794",
		Name:         "Aerial Maneuver",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1, "flying", "first strike")),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.PumpAndGrantFirstTarget(1, 1, "Aerial Maneuver — +1/+1, flying and first strike", "flying", "first strike")(ctx.Game, item)
		},
	})
}
