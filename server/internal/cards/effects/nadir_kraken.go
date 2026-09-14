package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nadir Kraken — Creature — Kraken {1}{U}{U}, 2/3 (EDHREC rank
// 2650):
//
//	"Whenever you draw a card, you may pay {1}. If you do, put a +1/+1
//	 counter on this creature and create a 1/1 blue Tentacle creature
//	 token."
//
// The draw-payoff that grows and goes wide for one mana a card. The
// trigger is the controller's own draw — one event per card, so a
// draw-three asks three times, as printed — and the payoff is the
// MayPay prompt for {1} (Mind's Eye's shape, Chooser the controller)
// with the counter and the Tentacle riding the "pay". A Kraken that
// left the battlefield before the payment takes no counter; the
// Tentacle still comes, since "if you do" is satisfied.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5a73d439-bd78-42e1-90ae-eedd30536881",
		Name:         "Nadir Kraken",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDrawCard},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Nadir Kraken — pay {1} for a +1/+1 counter and a Tentacle",
					func(g *game.Game, item *game.StackItem) error {
						return MayPay{
							Chooser:  item.Controller,
							Cost:     "{1}",
							Question: "Nadir Kraken — pay {1} to put a +1/+1 counter on it and create a 1/1 Tentacle?",
							OnPay: func(ctx *Context) error {
								return b25GrowAndSpawnTentacle(ctx, item.SourceCardID, item.Controller)
							},
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
