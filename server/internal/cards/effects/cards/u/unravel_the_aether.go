package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unravel the Aether — Instant {1}{G} (Born of the Gods, S82 OLD b377, #5559):
//
//	"Choose target artifact or enchantment. Its owner shuffles it into their
//	 library."
//
// The tuck-then-shuffle of Chaos Warp, to the OWNER's library; the shuffle is
// the tuck's continuation (S82OldB377ShuffleTargetIntoOwnersLibrary).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "128e0bd1-ed1d-4f53-8dae-83ee95f6fa2a",
		Name:         "Unravel the Aether",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.S82OldB377ShuffleTargetIntoOwnersLibrary(ctx.Game, item)
		},
	})
}
