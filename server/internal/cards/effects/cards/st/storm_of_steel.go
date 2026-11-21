package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Storm of Steel — {3}{R} Sorcery (Teenage Mutant Ninja Turtles
// Commander, #5011):
//
//	"Storm of Steel deals 2 damage to each of one or two targets."
//
// Prismari Charm's second bullet at 2: one or two any-targets (a creature,
// planeswalker, battle or player), each dealt 2 at once. One target can't
// be chosen twice (CR 115.3). A target that has become illegal takes
// nothing and the other still does (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d5021db4-64d3-4a0f-9167-499ccbe16e1e",
		Name:         "Storm of Steel",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny().WithCount(1, 2),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 2)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				ids = append(ids, t.ID)
			}
			return effects.DealDamageToEachThen(ctx, ids, 2, effects.RecipientThen{})
		},
	})
}
