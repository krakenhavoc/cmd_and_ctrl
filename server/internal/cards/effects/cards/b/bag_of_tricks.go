package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bag of Tricks — Artifact {1}{G} (Adventures in the Forgotten Realms
// Commander, S82):
//
//	"{4}{G}, {T}: Roll a d8. Reveal cards from the top of your library
//	 until you reveal a creature card with mana value equal to the result.
//	 Put that card onto the battlefield and the rest on the bottom of your
//	 library in a random order."
//
// The d8 goes through the one dice door; the tail reveals until a creature
// card whose mana value is the (natural) result, RevealUntilThenPutOntoBattlefield.
// With no such card the whole library is revealed and goes to the bottom in
// a random order, which is the printed outcome.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f14ce7d7-6cfa-4f88-a381-d6e912c18b5d",
		Name:         "Bag of Tricks",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{G}, {T}: Roll a d8, reveal until a creature card with that mana value, put it onto the battlefield",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{4}{G}"), effects.TapCost()),
			Effect:  effects.Do(effects.RollDice{Sides: 8, N: 1, Then: bagOfTricksReveal}),
		}},
	})
}

// bagOfTricksReveal is the rest of the ability, after the d8. An on-disk key.
var bagOfTricksReveal = effects.CardTail("dice/bag-of-tricks-reveal", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.Rolls) == 0 {
		return nil
	}
	result := effects.DiceResult(o)
	return effects.RevealUntilThenPutOntoBattlefield{
		Match:  func(c game.Card) bool { return c.IsCreature() && c.ManaValue() == result },
		Reason: "Bag of Tricks — revealed until a creature card with mana value equal to the roll",
	}.Apply(ctx)
})
