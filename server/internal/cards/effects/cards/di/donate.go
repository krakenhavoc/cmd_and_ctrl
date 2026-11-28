package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Donate — Sorcery {2}{U} (Urza's Destiny, S82 #5621):
//
//	"Target player gains control of target permanent you control."
//
// Harmless Offering's shape with any player as the first target (the caster may
// pick themselves, a gift to nobody). The gift has no stated duration, so it
// lasts for good (CR 611.2a) through a layer-2 control effect. It needs both
// targets legal on resolution (CR 608.2b), and a permanent that left and came
// back is a new object that isn't handed over (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "049d1fb2-cce7-40e7-a633-e84cf9782d89",
		Name:         "Donate",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetPlayer("target player"),
			effects.TargetPermanent("target permanent you control", effects.YouControl()),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			who, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			gift, ok := ctx.ClauseTarget(1)
			if !ok {
				return nil
			}
			return effects.GainControl{
				Target:     gift.ID,
				Controller: who.ID,
				Duration:   game.IndefiniteDuration(),
				Label:      "Donate — target player gains control of it",
			}.Apply(ctx)
		},
	})
}
