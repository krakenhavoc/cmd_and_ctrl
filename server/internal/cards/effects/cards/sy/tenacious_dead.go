package cardssy

import (
	"errors"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tenacious Dead — Creature — Skeleton Warrior {B}, 1/1 (Battlebond, #5529):
//
//	When this creature dies, you may pay {1}{B}. If you do, return it to the battlefield tapped under its owner's control.
//
// The offer is made only while the card is in a graveyard, and again when the
// payment finishes (a "dies" trigger looks back, CR 603.10a, so the card in the
// graveyard is the one that died); it returns through the entry pipeline tapped,
// under its owner's control (Ferocity of the Hunt's return).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9ae376d5-c68f-4a4b-851c-ebfd385b86dd",
		Name:         "Tenacious Dead",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Tenacious Dead — you may pay {1}{B}; if you do, return it to the battlefield tapped",
				func(g *game.Game, item *game.StackItem) error {
					if !tenaciousDeadInGraveyard(g, item) {
						return nil
					}
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}{B}",
						Question: "Tenacious Dead — pay {1}{B} to return it to the battlefield tapped?",
						Then:     tenaciousDeadPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

func tenaciousDeadInGraveyard(g *game.Game, item *game.StackItem) bool {
	z := g.FindCardZoneForEffect(item.SourceCardID)
	return z != nil && z.Kind == game.ZoneGraveyard
}

var tenaciousDeadPaid = effects.PayUnlessBranches("pay-unless/s82-old-b205-tenacious-dead-return",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		if !tenaciousDeadInGraveyard(ctx.Game, ctx.Item) {
			return nil
		}
		_, err := ctx.Game.ReturnToBattlefieldForEffect(ctx.Item.SourceCardID, uuid.Nil, true)
		if errors.Is(err, game.ErrCardNotFound) {
			return nil
		}
		return err
	}, nil)
