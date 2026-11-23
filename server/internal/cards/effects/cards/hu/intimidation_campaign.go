package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Intimidation Campaign — Enchantment {1}{U}{B} (Outlaws of Thunder
// Junction, #4468):
//
//	"When this enchantment enters, each opponent loses 1 life, you gain 1
//	 life, and you draw a card.
//	 Whenever you commit a crime, you may return this enchantment to its
//	 owner's hand. (It returns only from the battlefield. Targeting
//	 opponents, anything they control, and/or cards in their graveyards is
//	 a crime.)"
//
// The enters trigger is the drain and a draw in printed order. The crime
// trigger is the set's crime event (YouCommittedCrime) with a CR 603.5
// "may" asked as it resolves, returning the enchantment itself.
//
// No simplification.
func init() {
	const crime = "Intimidation Campaign — return this enchantment to its owner's hand"
	effects.Register(effects.Spec{
		OracleID:     "5915039b-3dbd-4a49-8bd3-23ff0fd9c90a",
		Name:         "Intimidation Campaign",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Intimidation Campaign — each opponent loses 1 life, you gain 1 life, you draw a card",
				effects.Do(effects.EachOpponentLosesLife{N: 1}, effects.GainLife{Amount: 1}, effects.DrawCards{N: 1})),
			effects.Optional(effects.On(game.EventBecomesTarget, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.YouCommittedCrime(ev, source)
			}, crime, func(g *game.Game, item *game.StackItem) error {
				return effects.BounceToHand{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			}), "Intimidation Campaign — return it to its owner's hand?"),
		},
	})
}
