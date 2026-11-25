package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nessian Boar — Creature — Boar {3}{G}{G}, 10/6 (Theros Beyond Death, S82 OLD,
// #5514):
//
//	"All creatures able to block this creature do so.
//	 Whenever this creature becomes blocked by a creature, that creature's
//	 controller draws a card."
//
// The first line is Breaker of Armies' block requirement (CR 509.1c): every
// creature that can block the Boar must. The second is one trigger per
// blocking creature (CR 509.3d), read off the block pair: it fires when the
// Boar is the attacker named by the pair, and its controller-to-be is the
// BLOCKER's controller as it stood when the block was declared, so a blocker
// that is gone by resolution still draws its controller a card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "139a5f0b-6f76-4918-96e6-cdcf27253230",
		Name:         "Nessian Boar",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.AllAbleToBlockDoSo()},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Target == source.InstanceID && ev.CardID != uuid.Nil
			}, "Nessian Boar — that creature's controller draws a card",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					ev := item.Trigger.Event
					controller := ev.Actor
					if c, ok := g.LookupCardForEffect(ev.CardID); ok {
						controller = c.Controller
					}
					if controller == uuid.Nil {
						return nil
					}
					return effects.DrawCards{Player: controller, N: 1}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
