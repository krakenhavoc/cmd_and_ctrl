package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mesa Enchantress — Creature — Human Druid {1}{W}{W}, 0/2 (EDHREC
// rank 811):
//
//	"Whenever you cast an enchantment spell, you may draw a card."
//
// The enchantress engine. Beast Whisperer's cast trigger with the
// type swapped and a "you may": the spell's type is read off the
// stack, and the draw resolves before the enchantment does (LIFO),
// as in paper.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8f4b8a19-72f4-48ed-ac05-62a7c7525797",
		Name:         "Mesa Enchantress",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && spell.IsEnchantment()
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Mesa Enchantress — draw a card?"},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Mesa Enchantress — draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
