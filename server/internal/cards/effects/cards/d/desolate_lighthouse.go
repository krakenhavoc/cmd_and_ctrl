package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Desolate Lighthouse — Land (EDHREC rank 2168):
//
//	"{T}: Add {C}.
//	 {1}{U}{R}, {T}: Draw a card, then discard a card."
//
// The Izzet looting land. The colorless tap is an ordinary mana
// ability; the loot is a CR 602 activated ability with a mana and a
// tap component — the two share the tap, so a Lighthouse tapped for
// mana cannot loot that turn, as printed. The body is lootOne: draw
// first, then the discard prompt, so the drawn card is a legal
// discard exactly as it is in paper.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aa6dbdf2-2379-4ff5-8a6c-70258784dc35",
		Name:         "Desolate Lighthouse",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}{R}, {T}: Draw a card, then discard a card",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}{U}{R}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.LootOne(g, item, 1)
			},
		}},
	})
}
