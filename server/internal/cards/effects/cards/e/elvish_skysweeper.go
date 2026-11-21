package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elvish Skysweeper — Creature — Elf Warrior {G}, 1/1 (Ravnica: City of
// Guilds, #5600):
//
//	"{4}{G}, Sacrifice a creature: Destroy target creature with flying."
//
// Bamboo Grove Archer's flying-only destroy on a sacrifice row. The
// Skysweeper can sacrifice itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f557697a-4c1b-44f6-a989-4879bbe26eb5",
		Name:         "Elvish Skysweeper",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{G}, Sacrifice a creature: Destroy target creature with flying.",
			Purpose: game.Purpose{Answers: game.AnswerRemove | game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{4}{G}"), effects.SacrificeACreature()),
			Targets: effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
