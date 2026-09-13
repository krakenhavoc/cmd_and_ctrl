package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flux Channeler — Creature — Human Wizard {2}{U}, 2/2:
//
//	"Whenever you cast a noncreature spell, proliferate."
//
// Magecraft's counter-themed cousin, and Storm-Kiln Artist's trigger
// shape with a wider net: NONCREATURE is every instant, sorcery,
// artifact, enchantment, planeswalker and land-that-is-somehow-cast,
// not just instants and sorceries.
//
// The trigger fires on CAST, so it goes on the stack ABOVE the spell
// that caused it and proliferates BEFORE that spell resolves. A
// proliferate that grows a planeswalker therefore happens before the
// spell you cast this turn resolves, which matters when the spell is
// the one that would have killed it.
func init() {
	Register(Spec{
		OracleID: "83874e60-291b-47a7-ba9f-69437fa7e3c7",
		Name:     "Flux Channeler",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && !spell.IsCreature()
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Flux Channeler — proliferate",
					func(g *game.Game, item *game.StackItem) error {
						return Proliferate{}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
