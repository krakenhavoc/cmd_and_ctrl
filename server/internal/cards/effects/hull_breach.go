package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hull Breach — Sorcery {R}{G} (EDHREC rank 2402):
//
//	"Choose one —
//	 • Destroy target artifact.
//	 • Destroy target enchantment.
//	 • Destroy target artifact and target enchantment."
//
// The Gruul two-for-one. Three modes, each with its own target
// clause; the first two are the plain shapes. The third names two
// "target" words with DIFFERENT predicates, and a TargetSpec carries
// one predicate for every slot — the per-slot predicate is the
// Bushwhack / Ram Through seam — so the third mode's picker accepts
// any two artifacts or enchantments and the resolution destroys the
// first artifact and the first enchantment among them.
//
// Declared simplification (weaker than printed, never stronger): a
// third-mode pick of two artifacts, or two enchantments, destroys
// only the first. An artifact enchantment picked for one slot fills
// the artifact slot (the first that matches wins), never both.
func init() {
	Register(Spec{
		OracleID:     "2da232d8-580f-4116-b977-2c59cd21b5a4",
		Name:         "Hull Breach",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The third mode's picker accepts any two artifacts or enchantments, but only the first artifact and the first enchantment you pick are destroyed — two of the same kind destroy just one."},
		Modes: ChooseOne(
			Mode("Destroy target artifact.", TargetPermanent("target artifact", Artifact())),
			Mode("Destroy target enchantment.", TargetPermanent("target enchantment", Enchantment())),
			Mode("Destroy target artifact and target enchantment.",
				TargetPermanent("target artifact and target enchantment", Or(Artifact(), Enchantment())).WithCount(2, 2)),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if ctx.HasMode(2) {
				return b22DestroyFirstArtifactAndFirstEnchantment(ctx)
			}
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					return DestroyTarget{Target: t.ID}.Apply(ctx)
				}
			}
			return nil
		},
	})
}
