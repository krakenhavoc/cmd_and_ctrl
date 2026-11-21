package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "a11d8372-87fb-40a2-b638-314c6a5af03b",
		Name:         "Break Down the Door",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.Mode("Exile target artifact.", effects.TargetPermanent("target artifact", effects.Artifact())),
			effects.Mode("Exile target enchantment.", effects.TargetPermanent("target enchantment", effects.Enchantment())),
			effects.Mode("Manifest dread."),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if ctx.HasMode(2) {
				return effects.ManifestDread{}.Apply(ctx)
			}
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if ctx.HasMode(0) || ctx.HasMode(1) {
				return effects.ExileTarget{Target: item.Targets[0].ID}.Apply(ctx)
			}
			return nil
		},
	})
}
