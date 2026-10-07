package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unnerving Grasp — Sorcery {2}{U}:
//
//	"Return up to one target nonland permanent to its owner's hand.
//	 Manifest dread."
//
// "Up to one" target: with no target chosen the spell still manifests
// dread. A target that became illegal in response is skipped, and the
// manifest happens anyway (CR 608.2b — the spell does as much as it
// can, and it only fizzles when every target is illegal).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7a335332-87ff-42dc-a547-c84295dd8138",
		Name:         "Unnerving Grasp",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("up to one target nonland permanent", Not(Land())).WithCount(0, 1),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			manifest := func(ctx *Context, _ bool) error { return ManifestDread{}.Apply(ctx) }
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					// The bounce can pause on a commander's CR 903.9
					// question, so "Manifest dread" waits behind it.
					return BounceToHand{Target: t.ID, Then: manifest}.Apply(ctx)
				}
			}
			return manifest(ctx, false)
		},
	})
}
