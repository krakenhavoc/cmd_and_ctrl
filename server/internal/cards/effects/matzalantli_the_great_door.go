package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Matzalantli, the Great Door // The Core — Legendary Artifact {3} //
// Legendary Land (Lost Caverns of Ixalan, #4470):
//
//	Front: "{T}: Draw a card, then discard a card.
//	        {4}, {T}: Transform Matzalantli. Activate only if there are four
//	        or more permanent types among cards in your graveyard."
//	Back:  "Fathomless descent — {T}: Add X mana of any one color, where X
//	        is the number of permanent cards in your graveyard."
//
// The transform is in place (CR 701.27, 712.18): the same object, untapped
// state unchanged. Its "only if" is the activation's Condition (CR 602.1b),
// read from the graveyard's printed types (permanentTypesInGraveyard); a card
// with two permanent types, an artifact creature, counts for both. The Core
// is a real mana ability (CR 605.1a: no target, adds mana), one pick of colour
// for all X (ProducedOneColor, as Empowered Autogenerator's), X read as the
// mana is produced.
//
// No simplification.
func init() {
	const oracle = "16182e01-22ff-4786-985d-919b47c4aa4d"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Matzalantli, the Great Door",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Draw a card, then discard a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return b16DrawThenDiscard(g, item, 1)
			},
		}, {
			Label: "{4}, {T}: Transform Matzalantli. Activate only if there are four or more permanent types among cards in your graveyard.",
			Cost:  Plus(ManaCost("{4}"), TapCost()),
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				return permanentTypesInGraveyard(g, controller) >= 4
			},
			Effect: Do(TransformThis{}),
		}},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "The Core",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{Tap: true},
			ProducedAmount: ProducedOneColor(func(g *game.Game, controller, _ uuid.UUID) int {
				return permanentCardsInGraveyard(g, controller)
			}),
			Label: "Fathomless descent — {T}: Add X mana of any one color, where X is the number of permanent cards in your graveyard",
		}},
	})
}
