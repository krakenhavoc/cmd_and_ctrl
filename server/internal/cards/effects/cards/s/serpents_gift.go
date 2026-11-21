package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Serpent's Gift — Instant {2}{G} (Magic 2013, #5568):
//
//	"Target creature gains deathtouch until end of turn. (Any amount of
//	 damage it deals to a creature is enough to destroy it.)"
//
// A layer 6 grant on the target, if still legal as it resolves (CR
// 608.2b), ending at cleanup (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b75fb789-6a78-4d23-b872-7627b478fb6f",
		Name:         "Serpent's Gift",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "deathtouch")),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.EbTargetCreatureGainsUntilEOT("deathtouch", "Serpent's Gift — deathtouch")(ctx.Game, item)
		},
	})
}
