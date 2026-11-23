package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fathom Mage — Creature — Human Wizard {2}{G}{U}, 1/1:
//
//	"Evolve (Whenever a creature you control enters, if that creature
//	 has greater power or toughness than this creature, put a +1/+1
//	 counter on this creature.)
//	 Whenever a +1/+1 counter is put on this creature, you may draw a
//	 card."
//
// #1841: the draw triggers once PER COUNTER, not once per placement
// (CR 603.2c; Fathom Mage's rulings). WheneverACounterIsPutOnThis reads
// the number of counters the placement event put, after Hardened Scales
// or Doubling Season has settled it, so two counters are two triggers
// and two "you may" prompts. Evolve is the engine's keyword trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "93d0e129-e3b5-4aff-9e50-f34771ed00ff",
		Name:            "Fathom Mage",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordEvolve},
		Triggered: []game.TriggeredAbility{
			effects.OptionalDraw(effects.WheneverACounterIsPutOnThis(game.CounterPlusOne, "Fathom Mage — you may draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
				}), "Fathom Mage — draw a card?", 1),
		},
	})
}
