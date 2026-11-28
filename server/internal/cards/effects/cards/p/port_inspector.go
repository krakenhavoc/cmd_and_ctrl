package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Port Inspector — Creature — Human {1}{U}, 1/2 (Mercadian Masques, S82 OLD batch 687, #5619):
//
//	"Whenever this creature becomes blocked, you may look at defending player's hand."
//
// Slate Street Ruffian's once-per-attacker "becomes blocked" event, whose actor is the
// defending player (CR 509.3c), optional at resolution (CR 603.5). Looking marks each
// card in that hand as known to the Inspector's controller alone, the way Glasses of
// Urza does, so no other player learns it; the hand is read as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "25450b06-3065-4b4f-975e-586d045e5815",
		Name:         "Port Inspector",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventBecomesBlocked, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventBecomesBlocked && ev.CardID == source.InstanceID
			}, "Port Inspector — you may look at defending player's hand",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					ctx := effects.NewContext(g, item)
					if p := ctx.PlayerByID(item.Trigger.Event.Actor); p != nil && p.Hand != nil {
						for i := range p.Hand.Cards {
							p.Hand.Cards[i].AddKnower(ctx.Controller())
						}
					}
					return nil
				}),
				"Port Inspector — look at defending player's hand?"),
		},
	})
}
