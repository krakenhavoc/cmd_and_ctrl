package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pollywog Symbiote — Creature — Frog {1}{U}, 1/3 (Ikoria: Lair of Behemoths, #5512):
//
//	"Each creature spell you cast costs {1} less to cast if it has mutate.
//	 Whenever you cast a creature spell, if it has mutate, draw a card, then
//	 discard a card."
//
// "Has mutate" is the keyword printed on the spell's card, read as it is
// cast; a creature spell cast for its mutate cost is still a creature spell
// with mutate. The "if" is on the spell cast and a printed keyword cannot
// change while the spell is on the stack, so it is asked once, as the spell
// is cast.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0a766368-c982-4cba-9638-c7a56978c240",
		Name:         "Pollywog Symbiote",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Each creature spell you cast costs {1} less to cast if it has mutate.",
				effects.YourSpell(), effects.CreatureSpell(), effects.SpellWithKeyword("mutate")),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.S82OldB115CreatureSpellWithMutateCastByYou, "Pollywog Symbiote — draw a card, then discard a card",
				func(g *game.Game, item *game.StackItem) error { return effects.LootOne(g, item, 1) }),
		},
	})
}
