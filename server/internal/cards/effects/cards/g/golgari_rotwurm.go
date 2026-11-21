package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "75df9e6a-a17c-4311-a4e7-ff0d75728337",
		Name:         "Golgari Rotwurm",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, Sacrifice a creature: Target player loses 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{B}"), effects.SacrificeACreature()),
			Targets: effects.TargetPlayer("target player"),
			Effect:  effects.Do(effects.TargetPlayerLosesLife{N: 1}),
		}},
	})
}
