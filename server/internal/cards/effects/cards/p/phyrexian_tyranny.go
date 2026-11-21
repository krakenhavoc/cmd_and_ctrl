package cardsp

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Phyrexian Tyranny — Enchantment {U}{B}{R} (S82 GAP, #5948):
//
//	"Whenever a player draws a card, that player loses 2 life unless they
//	 pay {2}."
//
// Elesh Norn's pay-or-lose: the drawing player is the payer (Chooser), read
// off the draw event when the trigger fires.
func init() {
	const label = "Phyrexian Tyranny — that player loses 2 life unless they pay {2}"
	effects.Register(effects.Spec{
		OracleID:     "55082cd0-a33c-4f1d-9f46-9578f63e6bb6",
		Name:         "Phyrexian Tyranny",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDrawCard},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor != uuid.Nil
			},
			Key: label,
			Effect: func(g *game.Game, item *game.StackItem) error {
				drawer := item.Trigger.Event.Actor
				if p := g.PlayerByIDForEffect(drawer); p == nil || p.Eliminated {
					return nil
				}
				return effects.PayUnless{
					Chooser:  drawer,
					Cost:     "{2}",
					Question: "Phyrexian Tyranny — pay {2}, or lose 2 life?",
					Else:     phyrexianTyrannyDecline,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

// phyrexianTyrannyDecline is the keyed decline: the player who drew loses 2
// life (#4256: a table waiting on the payer is a restore point).
var phyrexianTyrannyDecline = effects.PayUnlessElse("pay-unless/phyrexian-tyranny-lose-two", func(ctx *effects.Context, r game.PayUnlessAnswer) error {
	return ctx.Game.ChangePlayerLifeForEffect(r.Source, r.Chooser, -2)
})
