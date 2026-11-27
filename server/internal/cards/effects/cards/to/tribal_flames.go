package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tribal Flames — Sorcery {1}{R} (Invasion, #5616):
//
//	"Domain — Tribal Flames deals X damage to any target, where X is the number
//	 of basic land types among lands you control."
//
// X is read as the spell resolves (CR 608.2h), so a land that left in
// response counts for nothing. A target that is no longer legal leaves the
// spell with nothing to do (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c82eda61-d195-49c9-8ff0-a78dfa650689",
		Name:         "Tribal Flames",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			x := effects.DomainCount(ctx.Game, ctx.Controller())
			if x == 0 {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: x}.Apply(ctx)
		},
	})
}
