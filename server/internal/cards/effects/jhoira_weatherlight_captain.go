package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jhoira, Weatherlight Captain — Legendary Creature — Human Artificer
// {2}{U}{R}, 3/3 (EDHREC rank 1071):
//
//	"Whenever you cast a historic spell, draw a card. (Artifacts,
//	 legendaries, and Sagas are historic.)"
//
// The artifact-storm commander. "Historic" is CR 205.4h — an
// artifact, a legendary, or a Saga — read off the spell on the stack
// through b09IsHistoric. The trigger goes on the stack above the
// spell and resolves first, so the card is drawn before the historic
// spell lands, as in paper. Jhoira does not trigger on herself: she
// is on the stack, not the battlefield, when her own cast event
// fires.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c803b788-4213-4fab-b841-7e5bbf66088e",
		Name:         "Jhoira, Weatherlight Captain",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && b09IsHistoric(spell)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Jhoira, Weatherlight Captain — draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
