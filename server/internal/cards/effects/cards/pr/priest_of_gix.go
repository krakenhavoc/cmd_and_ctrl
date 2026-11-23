package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Priest of Gix — Creature — Phyrexian Human Cleric Minion {2}{B}, 2/1:
//
//	"When this creature enters, add {B}{B}{B}."
//
// A triggered ability that adds mana but triggers on an event other than
// a mana ability's activation, so it is not a mana ability (CR 605.1b):
// it uses the stack and the mana lands as the trigger resolves, to be
// spent in the same step.
//
// No simplification. Mass Casualty deck request (#3245).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d93f82ce-0eed-45cc-a7b1-50fd4cbb6152",
		Name:         "Priest of Gix",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Priest of Gix — add {B}{B}{B}", effects.Do(effects.AddMana{Produced: "{B}{B}{B}"})),
		},
	})
}
