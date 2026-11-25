package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Prince of Thralls — Creature — Demon {4}{U}{B}{B}{R}, 7/7 (Secret Lair Drop):
//
//	"Whenever a permanent an opponent controls is put into a graveyard, put
//	 that card onto the battlefield under your control unless that opponent
//	 pays 3 life."
//
// A permanent that went from the battlefield to a graveyard while an
// opponent of the Prince's controller controlled it, as it last existed
// (Athreos's reading of the dying permanent): any permanent type, but a token
// is gone and has no card to put. "That opponent" is its controller then. As it
// resolves that opponent either pays 3 life (CR 119.4) or the card, if it is
// still that card in the graveyard (CR 400.7), enters under the Prince's
// controller's control. A permanent that is put into a graveyard from anywhere
// else was not controlled by an opponent, so only the battlefield counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e7f3b4bb-ebf5-4e59-8f0b-aa26f78e9edc",
		Name:         "Prince of Thralls",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				if !ok || c.IsToken() {
					return false
				}
				who := effects.LeftUnderControlOf(ev, c)
				p := g.PlayerByIDForEffect(who)
				return who != uuid.Nil && who != source.Controller && p != nil && !p.Eliminated
			}, "Prince of Thralls — put that card onto the battlefield under your control unless that opponent pays 3 life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					ev := item.Trigger.Event
					c, ok := g.LookupCardForEffect(ev.CardID)
					if !ok {
						return nil
					}
					return princeOfThrallsAsk.Ask(ctx, effects.LeftUnderControlOf(ev, c), ev.CardID)
				}),
		},
	})
}

var princeOfThrallsAsk = effects.NewOldB015PayThreeLifeUnless("option-pick/prince-of-thralls-pay-or-take",
	"Prince of Thralls — let them put your card onto the battlefield, or pay 3 life?", "Let them take it",
	func(ctx *effects.Context, card uuid.UUID) error {
		if !effects.DiedCardStillInGraveyard(ctx, card) {
			return nil
		}
		return effects.ReturnFromGraveyard{Target: card, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}.Apply(ctx)
	})
