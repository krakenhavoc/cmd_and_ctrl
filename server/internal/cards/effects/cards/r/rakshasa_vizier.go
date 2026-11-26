package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rakshasa Vizier — Khans of Tarkir (S82 OLD b346, #5554):
//
//	Creature — Demon {2}{B}{G}{U}, 4/4: Whenever one or more cards are put into exile
//	from your graveyard, put that many +1/+1 counters on this creature.
//
// One trigger per batch of exiles; the count is the cards of the batch that
// left the controller's graveyard for exile (OldB346CardsExiledFromYourGraveyard).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9bdef081-1408-4125-82c3-04296dee05d5",
		Name:         "Rakshasa Vizier",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventZoneMove, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.OldB346CardExiledFromYourGraveyard(ev, source, g)
			}, "Rakshasa Vizier — put that many +1/+1 counters on this creature", func(g *game.Game, item *game.StackItem) error {
				n := effects.OldB346CardsExiledFromYourGraveyard(g, item)
				if n <= 0 {
					return nil
				}
				return effects.CounterOnThis{Kind: game.CounterPlusOne, N: n}.Apply(effects.NewContext(g, item))
			})),
		},
	})
}
