package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "d02befcb-3e2c-4cfe-a913-9bb648e5bb2c",
		Name:         "Aquitect's Will",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target land", Land()),
		OnResolve: ResolveSteps(
			// The draw waits for a prompt the counter asks (CR 608.2c, #4155).
			Instruction(func(ctx *Context) error { return FloodTargetLand(ctx, "Aquitect's Will") }),
			Instruction(func(ctx *Context) error {
				// ControlsAtLeast lends the spell's text, so "Merfolk" follows a
				// text change (CR 612.2, #3984).
				if !ControlsAtLeast(1, func(c game.Card) bool { return c.HasSubtypeWord("Merfolk") })(ctx.Game, ctx.Controller(), ctx.Source()) {
					return nil
				}
				return DrawCards{N: 1}.Apply(ctx)
			}),
		),
	})
}
