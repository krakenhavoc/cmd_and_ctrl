package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unlucky Drop — Instant {3}{U} (The Lost Caverns of Ixalan, #4470):
//
//	"Target artifact or creature's owner puts it on their choice of the
//	 top or bottom of their library."
//
// Run Behind's placement without the discount: the choice is the OWNER's,
// not the caster's, so the top-or-bottom prompt goes to the player who
// owns the permanent (ownerTucksFirstTargetTopOrBottom). A target that is
// gone or no longer an artifact or creature makes the spell do nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "98ebde8c-b47d-440b-93cb-a1407b807074",
		Name:         "Unlucky Drop",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.OwnerTucksFirstTargetTopOrBottom(ctx, "Unlucky Drop — put it on your choice of the top or bottom of your library")
		},
	})
}
