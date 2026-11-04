package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "c3c25a2c-f71c-4cb7-9117-0bcda9248bfd",
		Name:         "Prismari Charm",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(Mode("Surveil 2, then draw a card."), game.Purpose{Draws: 1}),
			ModeWithPurpose(Mode("Prismari Charm deals 1 damage to each of one or two targets.",
				TargetAny().WithCount(1, 2)), ForTargets(DamageToTarget(0, 1))),
			Mode("Return target nonland permanent to its owner's hand.",
				TargetPermanent("target nonland permanent", Nonland())),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			switch {
			case ctx.HasMode(0):
				return Surveil{Player: item.Controller, N: 2, ThenTail: ctx.Tail(thenDrawACard)}.Apply(ctx)
			case ctx.HasMode(1):
				var ids []uuid.UUID
				for _, t := range ctx.LegalTargets() {
					ids = append(ids, t.ID)
				}
				return DealDamageToEachThen(ctx, ids, 1, RecipientThen{})
			case ctx.HasMode(2):
				return bounceChosenTarget(ctx.Game, item)
			}
			return nil
		},
	})
}
