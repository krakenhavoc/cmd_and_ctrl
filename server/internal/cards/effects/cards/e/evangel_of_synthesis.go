package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Evangel of Synthesis — Creature — Phyrexian Human Cleric {U}{B}, 2/3 (The
// Brothers' War, #5482):
//
//	"When this creature enters, draw a card, then discard a card.
//	 As long as you've drawn two or more cards this turn, this creature
//	 gets +1/+0 and has menace."
//
// The loot is Adventurer's Airship's draw-then-discard. The pump is a layer
// 7c modify and the menace a layer 6 grant, both on itself and both read
// off the turn's draw tally (Messenger Hawk, Foggy Swamp Hunters), so the
// card drawn by the loot itself counts toward the second.
//
// No simplification.
func init() {
	pump := game.StaticAbility{
		Label:     "Gets +1/+0 as long as you've drawn two or more cards this turn.",
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: effects.S82BroB07DrewTwoOrMore,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power++
		},
		DependsOnHandSize: true,
	}
	effects.Register(effects.Spec{
		OracleID:     "dc0bc051-2d5c-4b7f-9902-4f9d5ef8f19d",
		Name:         "Evangel of Synthesis",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{pump, effects.S82BroB07SelfKeywordWhileDrawnTwo("menace")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Evangel of Synthesis — draw a card, then discard a card",
				func(g *game.Game, item *game.StackItem) error { return effects.B16DrawThenDiscard(g, item, 1) }),
		},
	})
}
