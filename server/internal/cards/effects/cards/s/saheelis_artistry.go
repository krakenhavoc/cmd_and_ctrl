package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Saheeli's Artistry — Sorcery {4}{U}{U}:
//
//	"Choose one or both —
//	 • Create a token that's a copy of target artifact.
//	 • Create a token that's a copy of target creature, except it's an
//	   artifact in addition to its other types."
//
// Azula Always Lies' "one or both" and Mirage Mockery's token copies off
// each mode's own target. The second bullet's exception is Molten
// Duplication's (CR 707.9b). A bullet whose target is illegal as the
// spell resolves does nothing, and the other still happens (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0038e74a-b950-47ff-8853-dba6a3b31091",
		Name:         "Saheeli's Artistry",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose one or both", 1, 2,
			effects.ModeDoing("Create a token that's a copy of target artifact.",
				effects.TargetPermanent("target artifact", effects.Artifact()),
				effects.TokenCopyTheModesTarget),
			effects.ModeDoing("Create a token that's a copy of target creature, except it's an artifact in addition to its other types.",
				effects.TargetCreature("target creature"),
				func(item *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return effects.CreateTokenCopy{Controller: item.Controller, Copy: t.ID, N: 1, Except: effects.TokenCopyIsAlsoAnArtifact}.Apply(ctx)
				}),
		),
	})
}
