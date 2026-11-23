package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Smaug the Impenetrable — Legendary Creature — Dragon {5}{B}{R}, 8/7 (The
// Hobbit Eternal):
//
//	"Flying, indestructible, haste
//	 Whenever Smaug is dealt noncombat damage, create that many Treasure
//	 tokens."
//
// Three keywords on PrintedKeywords and one trigger on the damage event
// (Screaming Nemesis's condition, b35SelfWasDealtDamage, narrowed to damage
// that is not combat damage, CR 510). The amount is the event's: the damage
// actually dealt, so damage prevented does not count, and it is read from the
// triggering event as the ability resolves. Being indestructible, Smaug stays
// to collect however much lands on him.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ce2fc562-7a71-4846-bfef-0a7c5f6c1ff3",
		Name:            "Smaug the Impenetrable",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "indestructible", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return !ev.Combat && effects.B35SelfWasDealtDamage(ev, source)
			}, "Smaug the Impenetrable — create that many Treasure tokens",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil || item.Trigger.Event.Amount <= 0 {
						return nil
					}
					return effects.CreateToken{Controller: item.Controller, Template: effects.TreasureToken(), N: item.Trigger.Event.Amount}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
