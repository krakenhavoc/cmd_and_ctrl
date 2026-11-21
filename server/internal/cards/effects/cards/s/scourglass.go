package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scourglass — Artifact {3}{W}{W} (Marvel Super Heroes Commander,
// #4455):
//
//	"{T}, Sacrifice this artifact: Destroy all permanents except for
//	 artifacts and lands. Activate only during your upkeep."
//
// Nevinyrral's Disk's sweep with the class inverted: every permanent that
// is neither an artifact nor a land (creatures, enchantments,
// planeswalkers, battles) is destroyed together through the destroy
// pipeline, so indestructible permanents survive and a regeneration
// shield still works. The tap and the sacrifice are costs paid at announce.
// DuringYourUpkeep is the step AND your turn (Hell's Caretaker).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "78541a40-c27a-49c0-aa3e-5a48b97065f0",
		Name:         "Scourglass",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "{T}, Sacrifice this artifact: Destroy all permanents except for artifacts and lands. Activate only during your upkeep.",
			Purpose:   game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, Partial: true}},
			Cost:      effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Condition: effects.DuringYourUpkeep(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DestroyAllMatching{Match: effects.Not(effects.Or(effects.Artifact(), effects.Land()))}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
