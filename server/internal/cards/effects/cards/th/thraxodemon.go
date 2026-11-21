package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thraxodemon — Creature — Demon {1}{B}, 2/2 (The Brothers' War, S82):
//
//	"{3}, {T}, Sacrifice another creature or artifact: Draw a card."
//
// The sacrifice is a cost (CR 602.2), never the Thraxodemon itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "01cf689c-5a64-4ecf-b45d-3f385c9b212b",
		Name:         "Thraxodemon",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}, Sacrifice another creature or artifact: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, Draws: 1},
			Cost: effects.Plus(effects.ManaCost("{3}"), effects.TapCost(),
				effects.SacrificeAnotherN(1, "another creature or artifact", effects.Or(effects.Creature(), effects.Artifact()))),
			Effect: effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
