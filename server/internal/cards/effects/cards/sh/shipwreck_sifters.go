package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shipwreck Sifters — Creature — Spirit {1}{U}, 1/2 (Innistrad: Midnight
// Hunt, S82):
//
//	"When this creature enters, draw a card, then discard a card.
//	 Whenever you discard a Spirit card or a card with disturb, put a +1/+1
//	 counter on this creature."
//
// The loot is the shared draw-then-discard, so the discard it makes sets
// off the second ability. That one fires once per card you discard. The
// card is judged as it was in the hand (CR 603.10a): a Spirit card by its
// front face's types, a card with disturb by its keyword (CR 702.146a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d70ee6f1-541b-43b4-897c-68a231beb894",
		Name:         "Shipwreck Sifters",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Shipwreck Sifters — draw a card, then discard a card",
				func(g *game.Game, item *game.StackItem) error { return effects.LootOne(g, item, 1) }),
			effects.On(game.EventDiscardCard, effects.YouDiscardedA(func(c game.Card) bool {
				return effects.WasCardInAGraveyard(c, func(x game.Card) bool { return x.HasSubtypeWord("Spirit") }) || effects.HasFlashbackOrDisturb(c)
			}), "Shipwreck Sifters — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
