package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Break Down the Door — Instant {2}{G}:
//
//	"Choose one —
//	 • Exile target artifact.
//	 • Exile target enchantment.
//	 • Manifest dread."
//
// The first two modes target and the third does not; the engine applies
// the chosen mode's target clause only, so choosing manifest dread
// asks for no target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a11d8372-87fb-40a2-b638-314c6a5af03b",
		Name:         "Break Down the Door",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Exile target artifact.", TargetPermanent("target artifact", Artifact())),
			Mode("Exile target enchantment.", TargetPermanent("target enchantment", Enchantment())),
			Mode("Manifest dread."),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if ctx.HasMode(2) {
				return ManifestDread{}.Apply(ctx)
			}
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if ctx.HasMode(0) || ctx.HasMode(1) {
				return ExileTarget{Target: item.Targets[0].ID}.Apply(ctx)
			}
			return nil
		},
	})
}
