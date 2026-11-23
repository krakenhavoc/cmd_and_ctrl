package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "7a335332-87ff-42dc-a547-c84295dd8138",
		Name:         "Unnerving Grasp",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("up to one target nonland permanent", effects.Not(effects.Land())).WithCount(0, 1),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					// The bounce can pause on a commander's CR 903.9
					// question, so "Manifest dread" waits behind it.
					return effects.BounceToHand{Target: t.ID, ThenTail: ctx.ItemTail(effects.ThenManifestDread)}.Apply(ctx)
				}
			}
			return effects.ManifestDread{}.Apply(ctx)
		},
	})
}
