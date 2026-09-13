package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wear Down — Sorcery {1}{G} (EDHREC rank 2403):
//
//	"Gift a card (You may promise an opponent a gift as you cast this
//	 spell. If you do, they draw a card before its other effects.)
//	 Destroy target artifact or enchantment. If the gift was
//	 promised, instead destroy two target artifacts and/or
//	 enchantments."
//
// A Naturalize that can become a two-for-one. The base mode is the
// single artifact-or-enchantment target, the shared shape.
//
// Sandbox simplification, declared — Starfall Invocation's posture:
// the GIFT is not offered. Gift is a cast-time promise (CR 702.174)
// that needs a prompt in the cast flow and a per-cast flag the
// resolution reads, and neither exists; the promised branch would
// also need a target clause that changes with the promise. So the
// spell is always its base mode, which is the printed card with one
// option removed — weaker and never stronger.
func init() {
	Register(Spec{
		OracleID:     "27905301-333e-4cdd-90cf-188159fcf8e9",
		Name:         "Wear Down",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The gift can't be promised, so the spell always destroys one target artifact or enchantment — never two."},
		Targets:      TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					return DestroyTarget{Target: t.ID}.Apply(ctx)
				}
			}
			return nil
		},
	})
}
