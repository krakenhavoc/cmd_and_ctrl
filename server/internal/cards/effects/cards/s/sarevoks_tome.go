package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sarevok's Tome — Artifact — Book {4} (Commander Legends: Battle for
// Baldur's Gate, #5948):
//
//	"When this artifact enters, you take the initiative.
//	 {T}: Add {C}. If you have the initiative, add {C}{C} instead.
//	 {3}, {T}: Exile cards from the top of your library until you exile a
//	 nonland card. You may cast that card without paying its mana cost.
//	 Activate only if you've completed a dungeon."
//
// The mana is read as the ability resolves (CR 605.3b), so it follows the
// initiative as it is then. The third ability is Nico Minoru's exile-until-
// nonland free cast, gated at activation on having completed a dungeon
// (CR 309.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6ec54d84-7026-45a3-a7d7-37cd5c9a8463",
		Name:         "Sarevok's Tome",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouTakeTheInitiative("Sarevok's Tome"),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost: effects.ManaAbilityCost{Tap: true},
			ProducedFunc: func(g *game.Game, controller, _ uuid.UUID) string {
				if effects.YouHaveTheInitiative(g, controller) {
					return "{C}{C}"
				}
				return "{C}"
			},
			Label: "{T}: Add {C}. If you have the initiative, add {C}{C} instead.",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Exile cards from the top of your library until you exile a nonland card. You may cast that card without paying its mana cost. Activate only if you've completed a dungeon.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				return effects.YouCompletedADungeon(g, controller)
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ExileUntilNonlandCastFree(effects.NewContext(g, item), "Sarevok's Tome")
			},
		}},
	})
}
