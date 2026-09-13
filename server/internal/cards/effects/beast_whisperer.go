package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Beast Whisperer — 2/3 Creature — Elf Druid for {2}{G}{G}:
//
//	"Whenever you cast a creature spell, draw a card."
//
// S19 sub-PR 6: the simplest cast trigger — your own creature
// spells only. AppliesTo looks the cast card up on the stack
// (EventCast.CardID) to read its type; the Whisperer itself
// doesn't trigger on its own cast because it's not on the
// battlefield yet when its EventCast fires. Mandatory; the draw
// happens when the trigger resolves, so it resolves BEFORE the
// creature spell that caused it (LIFO), as in paper.
func init() {
	Register(Spec{
		OracleID:     "5da7eea8-bb9e-47ce-a554-8a1ee058bd7a",
		Name:         "Beast Whisperer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && spell.IsCreature()
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Beast Whisperer — draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
