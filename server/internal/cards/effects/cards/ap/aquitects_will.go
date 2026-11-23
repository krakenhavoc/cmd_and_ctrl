package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aquitect's Will — Kindred Sorcery — Merfolk for {U}:
//
//	"Put a flood counter on target land. That land is an Island in addition to its other types for as long as it has a flood counter on it. If you control a Merfolk, draw a card."
//
// ADR 0109 §2 (#1604): the land is an Island in addition to its other
// types (CR 205.1b, an addSubtypes record) for as long as it has a
// flood counter on it (game.WhilePinnedHasCounter). Removing the last
// flood counter ends it for good (CR 611.2b). "If you control a
// Merfolk" is asked as the spell resolves, after the land has become
// an Island.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d02befcb-3e2c-4cfe-a913-9bb648e5bb2c",
		Name:         "Aquitect's Will",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target land", effects.Land()),
		OnResolve: effects.ResolveSteps(
			// The draw waits for a prompt the counter asks (CR 608.2c, #4155).
			effects.Instruction(func(ctx *effects.Context) error { return effects.FloodTargetLand(ctx, "Aquitect's Will") }),
			effects.Instruction(func(ctx *effects.Context) error {
				// ControlsAtLeast lends the spell's text, so "Merfolk" follows a
				// text change (CR 612.2, #3984).
				if !effects.ControlsAtLeast(1, func(c game.Card) bool { return c.HasSubtypeWord("Merfolk") })(ctx.Game, ctx.Controller(), ctx.Source()) {
					return nil
				}
				return effects.DrawCards{N: 1}.Apply(ctx)
			}),
		),
	})
}
