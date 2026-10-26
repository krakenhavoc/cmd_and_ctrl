package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Beacon of Unrest — Sorcery {3}{B}{B}:
//
//	"Put target artifact or creature card from a graveyard onto the
//	 battlefield under your control. Shuffle Beacon of Unrest into its
//	 owner's library."
//
// The reanimation family's body (reanimateSingleTarget) over any
// graveyard, then Beacon of Immortality's tail: the spell tucks itself
// into its owner's library and that library is shuffled, instead of
// going to the graveyard (TuckToLibraryThenForEffect). With its one
// target gone the spell does nothing at all (CR 608.2b), so it is not
// shuffled away either.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aef874fa-93c0-4f86-8b0d-dc24ec51467d",
		Name:         "Beacon of Unrest",
		Completeness: CompletenessFull,
		Targets: TargetCardInGraveyard("target artifact or creature card from a graveyard",
			Or(Artifact(), Creature())),
		Purpose: ForTargets(Reanimates(0)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			reanimateSingleTarget(ctx, ctx.Controller())
			owner := item.Owner
			return ctx.Game.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{ctx.Source()}, game.TuckOptions{}, game.KeyedTail{Ref: thenShuffle, Player: owner})
		},
	})
}
