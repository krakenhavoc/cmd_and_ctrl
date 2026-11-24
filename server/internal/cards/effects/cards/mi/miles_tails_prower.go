package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Miles "Tails" Prower — Legendary Creature — Fox Artificer {1}{W}{U}, 3/1
// (Secret Lair Drop):
//
//	"Flash
//	 Flying
//	 Whenever a Vehicle you control enters, draw a card if it has flying.
//	 Otherwise, put a flying counter on it."
//
// The entering Vehicle's flying is read through the layered view as the
// trigger resolves (CR 603.4 has no "if" here: both branches are the
// effect). A Vehicle that has left by then gets neither (CR 608.2b).
func init() {
	const label = "Miles \"Tails\" Prower — draw a card if it has flying, otherwise put a flying counter on it"
	effects.Register(effects.Spec{
		OracleID:        "98dd8231-5679-404a-b976-0bbfa6a8493c",
		Name:            "Miles \"Tails\" Prower",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
				return ok && c.HasSubtype("Vehicle")
			}, label, func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil || !effects.OnBattlefield(g, item.Trigger.Event.CardID) {
					return nil
				}
				id := item.Trigger.Event.CardID
				c, ok := g.LookupCardForEffect(id)
				if !ok {
					return nil
				}
				ctx := effects.NewContext(g, item)
				if game.HasKeyword(&c, "flying") {
					return effects.DrawCards{N: 1}.Apply(ctx)
				}
				return effects.AddCounter{Target: id, Kind: game.CounterFlying, N: 1}.Apply(ctx)
			}),
		},
	})
}
