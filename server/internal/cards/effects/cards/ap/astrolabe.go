package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Astrolabe — Artifact {3} (Alliances, #5634):
//
//	"{1}, {T}, Sacrifice this artifact: Add two mana of any one color. Draw a
//	 card at the beginning of the next turn's upkeep."
//
// A mana ability (CR 605.1a: it can add mana, has no target, and is not a
// loyalty ability) whose cost is {1}, tap and sacrifice. Its second sentence
// is the mana ability's rider: a delayed trigger for the next upkeep at the
// table (CR 603.7a), scheduled as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d19567ac-1f52-40eb-b2cf-4c791823ec92",
		Name:         "Astrolabe",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true, Sacrifice: true, Mana: "{1}"},
			Produced: effects.OneColorOfAmount(2),
			Label:    "{1}, {T}, Sacrifice this artifact: Add two mana of any one color. Draw a card at the beginning of the next turn's upkeep.",
			Answers:  game.AnswerSacOutlet,
			Rider: func(g *game.Game, controller, source uuid.UUID) error {
				return effects.DrawACardAtTheNextUpkeep("Astrolabe — draw a card").Apply(effects.KeyedContext(g, controller, source))
			},
		}},
	})
}
