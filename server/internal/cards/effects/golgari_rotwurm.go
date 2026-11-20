package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Golgari Rotwurm — Creature — Zombie Wurm {3}{B}{G}, 5/4 (Ravnica: City of
// Guilds, #5600):
//
//	"{B}, Sacrifice a creature: Target player loses 1 life."
//
// Gnawing Zombie's sacrifice outlet without the gain: the target player
// loses 1 life. The Rotwurm can sacrifice itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "75df9e6a-a17c-4311-a4e7-ff0d75728337",
		Name:         "Golgari Rotwurm",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{B}, Sacrifice a creature: Target player loses 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{B}"), SacrificeACreature()),
			Targets: TargetPlayer("target player"),
			Effect:  Do(TargetPlayerLosesLife{N: 1}),
		}},
	})
}
