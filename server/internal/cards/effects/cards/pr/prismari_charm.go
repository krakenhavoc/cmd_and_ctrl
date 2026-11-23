package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Prismari Charm — Instant {U}{R} (Secrets of Strixhaven, #4456):
//
//	"Choose one —
//	 • Surveil 2, then draw a card.
//	 • Prismari Charm deals 1 damage to each of one or two targets.
//	 • Return target nonland permanent to its owner's hand."
//
// Mode 0 is Preordain's shape: the draw rides the surveil's continuation so the
// card drawn comes from the library the player arranged (Risky Research). Mode 1
// takes one or two any-targets (a creature, planeswalker, battle or player), each
// dealt 1 damage at once; a target that has become illegal takes nothing and the
// other still does (CR 608.2b). Mode 2 is the usual bounce of a nonland
// permanent to its owner's hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c3c25a2c-f71c-4cb7-9117-0bcda9248bfd",
		Name:         "Prismari Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.Mode("Surveil 2, then draw a card."), game.Purpose{Draws: 1}),
			effects.ModeWithPurpose(effects.Mode("Prismari Charm deals 1 damage to each of one or two targets.",
				effects.TargetAny().WithCount(1, 2)), effects.ForTargets(effects.DamageToTarget(0, 1))),
			effects.Mode("Return target nonland permanent to its owner's hand.",
				effects.TargetPermanent("target nonland permanent", effects.Nonland())),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			switch {
			case ctx.HasMode(0):
				return effects.Surveil{Player: item.Controller, N: 2, ThenTail: ctx.Tail(effects.ThenDrawACard)}.Apply(ctx)
			case ctx.HasMode(1):
				var ids []uuid.UUID
				for _, t := range ctx.LegalTargets() {
					ids = append(ids, t.ID)
				}
				return effects.DealDamageToEachThen(ctx, ids, 1, effects.RecipientThen{})
			case ctx.HasMode(2):
				return effects.BounceChosenTarget(ctx.Game, item)
			}
			return nil
		},
	})
}
