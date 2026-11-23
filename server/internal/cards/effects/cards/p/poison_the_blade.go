package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Poison the Blade — Instant {1}{G} (Commander Legends: Battle for Baldur's
// Gate, #5948):
//
//	"Target creature gains deathtouch until end of turn.
//	 Draw a card."
//
// The deathtouch is an until-end-of-turn keyword on the chosen creature, and
// the draw follows as its own step (CR 608.2c): the spell does both, or with
// its target gone, neither (CR 608.2b: it is countered on resolution).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8af22295-0d6b-4cdc-b815-eb81ade241ec",
		Name:         "Poison the Blade",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Targets: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "deathtouch")).Targets},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"deathtouch"}, Label: "Poison the Blade — deathtouch"}.Apply(ctx)
			}),
			effects.DrawCards{N: 1},
		),
	})
}
