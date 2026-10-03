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
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := FloodTargetLand(ctx, "Aquitect's Will"); err != nil {
				return err
			}
			if !ControlsAtLeast(1, func(c game.Card) bool { return c.HasSubtype("Merfolk") })(ctx.Game, ctx.Controller(), ctx.Source()) {
				return nil
			}
			return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
