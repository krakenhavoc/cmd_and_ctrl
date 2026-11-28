package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Metalworker — Artifact Creature — Construct {3}, 1/2 (Urza's Destiny, S82
// #5621):
//
//	"{T}: Reveal any number of artifact cards in your hand. Add {C}{C} for each
//	 card revealed this way."
//
// Sacellum Godspeaker's mana ability that reveals any number of matching hand
// cards, none included: they are revealed as the ability resolves (CR 605.3b,
// CR 701.20a), stay in the hand, and each adds {C}{C}. The quality is artifact
// cards (game.RevealCost.CardType).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9ed9accb-9f0c-40fa-aeea-2d1e1c9d1bd1",
		Name:         "Metalworker",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost: effects.ManaAbilityCost{
				Tap:    true,
				Reveal: effects.RevealAnyNumber("artifact cards", game.RevealCost{CardType: "artifact"}),
			},
			ProducedForPaid: metalworkerProduced,
			Label:           "{T}: Reveal any number of artifact cards in your hand. Add {C}{C} for each card revealed this way.",
		}},
	})
}

// metalworkerProduced is "Add {C}{C} for each card revealed this way".
func metalworkerProduced(_ *game.Game, _, _ uuid.UUID, paid game.PaidCost) game.ManaOutput {
	return game.ManaOutput{Amount: game.ManaAmount{Options: []string{"C"}, N: 2 * len(paid.Revealed)}}
}
