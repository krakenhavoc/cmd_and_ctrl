package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Enrage — Instant {X}{R} (Scourge):
//
//	"Target creature gets +X/+0 until end of turn."
//
// Hatred's +X/+0 with X paid in mana instead of life: X is announced with
// the cast (CR 601.2b) and read as the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0f6e66d5-4f27-485b-999d-ee55c1e218b9",
		Name:         "Enrage",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      effects.ForTargets(effects.PowerXUntilEOT(0)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.BoostUntilEOT{Target: id, Power: ctx.X(), Label: "Enrage — +X/+0 until end of turn"}.Apply(ctx)
		},
	})
}
