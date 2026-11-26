package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skeletonize — {4}{R} Instant (Shards of Alara, S82 OLD b516):
//
//	"Skeletonize deals 3 damage to target creature. When a creature dealt
//	 damage this way dies this turn, create a 1/1 black Skeleton creature
//	 token with '{B}: Regenerate this token.'"
//
// Reckless Blaze's per-recipient rider on one recipient: a creature actually
// dealt damage (a prevented blow is not "dealt damage this way") gets a delayed
// "when it dies this turn" about it as the object it is now (CR 603.7a, 603.7c),
// and the Skeleton is created under the spell's controller. The Skeleton's
// regenerate ability is the token's own (the printed token reads "this
// creature").
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2a4cde8f-38e0-439b-9928-95ab23317f70",
		Name:         "Skeletonize",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || !ctx.IsTargetLegal(t) {
				return nil
			}
			return effects.DealDamageToEachThen(ctx, []uuid.UUID{t.ID}, 3, effects.ItemRecipient(item, skeletonizeDealt))
		},
	})
}

var (
	// skeletonizeDealt is the per-recipient rider: a creature dealt more than
	// 0 damage gets the delayed death watch. An on-disk key.
	skeletonizeDealt = effects.RecipientTail("recipient/old-b516-skeletonize-watch-death", func(ctx *effects.Context, target uuid.UUID, dealt int, _ game.TailOutcome) error {
		if dealt <= 0 {
			return nil
		}
		return effects.WhenThatCreatureDiesThisTurn(ctx, target, "Skeletonize — create a 1/1 black Skeleton creature token with \"{B}: Regenerate this token.\"", skeletonizeSkeleton)
	})
	// skeletonizeSkeleton is the fired trigger's body. An on-disk key.
	skeletonizeSkeleton = game.SimpleDelayedBody("old-b516-skeletonize/skeleton", func(g *game.Game, item *game.StackItem) error {
		return effects.CreateToken{Template: effects.OldB516SkeletonToken(), N: 1}.Apply(effects.NewContext(g, item))
	})
)
