package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Soul-Shackled Zombie — Creature — Zombie {3}{B}, 4/2:
//
//	"When this creature enters, exile up to two target cards from a
//	 single graveyard. If at least one creature card was exiled this
//	 way, each opponent loses 2 life and you gain 2 life."
//
// #1807, ADR 0106 §5. "Exiled this way" is what ACTUALLY reached exile
// (#870): the type is read before the move and the drain is gated on
// the batch's continuation, so a creature card whose exile was
// replaced, or a commander its owner sent to the command zone, does
// not pay out.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f81ca3a5-356d-4b5d-afc2-a33aa774816b",
		Name:         "Soul-Shackled Zombie",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersExileFromASingleGraveyard("Soul-Shackled Zombie", 2, func(g *game.Game, item *game.StackItem) error {
				return exileTargetCardsThen(NewContext(g, item), func(ctx *Context, exiled []uuid.UUID, wasCreature map[uuid.UUID]bool) error {
					if !anyWasCreature(exiled, wasCreature) {
						return nil
					}
					if err := eachOpponentLosesLife(ctx.Game, ctx.Item, 2); err != nil {
						return err
					}
					return GainLife{Player: ctx.Controller(), Amount: 2}.Apply(ctx)
				})
			}),
		},
	})
}
