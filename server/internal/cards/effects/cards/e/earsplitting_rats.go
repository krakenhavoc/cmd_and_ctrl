package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Earsplitting Rats — Creature — Rat {3}{B}, 2/1 (Judgment):
//
//	"When this creature enters, each player discards a card.
//	 Discard a card: Regenerate this creature."
//
// Kardur's Vicious Return's each-player discard (every player chooses, active
// player first) on the enters trigger, and Mortivore's regeneration with a
// discard cost instead of mana.
//
// No simplification.
func init() {
	const name = "Earsplitting Rats"
	effects.Register(effects.Spec{
		OracleID:     "6c14ecbf-0097-4fd8-97a3-a32b532c21e0",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(name+" — each player discards a card", effects.TspB1EachPlayerDiscardsACard(name+" — discard a card")),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Discard a card: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.DiscardACard(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Regenerate{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
