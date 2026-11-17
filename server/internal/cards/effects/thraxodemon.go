package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thraxodemon — Creature — Demon {1}{B}, 2/2 (The Brothers' War, S82):
//
//	"{3}, {T}, Sacrifice another creature or artifact: Draw a card."
//
// The sacrifice is a cost (CR 602.2), never the Thraxodemon itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "01cf689c-5a64-4ecf-b45d-3f385c9b212b",
		Name:         "Thraxodemon",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}, Sacrifice another creature or artifact: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, Draws: 1},
			Cost: Plus(ManaCost("{3}"), TapCost(),
				SacrificeAnotherN(1, "another creature or artifact", Or(Creature(), Artifact()))),
			Effect: Do(DrawCards{N: 1}),
		}},
	})
}
