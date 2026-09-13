package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "aa6dbdf2-2379-4ff5-8a6c-70258784dc35",
		Name:         "Desolate Lighthouse",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{1}{U}{R}, {T}: Draw a card, then discard a card",
			Cost:  Plus(ManaCost("{1}{U}{R}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return lootOne(g, item, 1)
			},
		}},
	})
}
