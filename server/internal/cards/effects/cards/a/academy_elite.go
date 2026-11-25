package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Academy Elite — Creature — Human Wizard {3}{U}, 0/0 (Commander 2016, S82
// older sets, #5537):
//
//	"This creature enters with X +1/+1 counters on it, where X is the number
//	 of instant and sorcery cards in all graveyards.
//	 {2}{U}, Remove a +1/+1 counter from this creature: Draw a card, then
//	 discard a card."
//
// X is counted as the creature enters, inside the replacement (CR 614.1c), so
// a card put into a graveyard in response is counted. The loot is one
// sentence: the discard waits for the prompt the draw may ask (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ba6c3c72-c014-45c6-a0b4-59eb9a65303e",
		Name:         "Academy Elite",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B19EntersWithCountersCounted(game.CounterPlusOne,
				func(g *game.Game, _ *game.Card) int { return effects.S82OldB256InstantSorceryCardsInAllGraveyards(g) },
				"Academy Elite enters with X +1/+1 counters on it, where X is the number of instant and sorcery cards in all graveyards"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}, Remove a +1/+1 counter from this creature: Draw a card, then discard a card.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{U}"), effects.RemoveCountersFromThis(game.CounterPlusOne, 1)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.B16DrawThenDiscard(g, item, 1)
			},
		}},
	})
}
