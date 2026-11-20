package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "f557697a-4c1b-44f6-a989-4879bbe26eb5",
		Name:         "Elvish Skysweeper",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{4}{G}, Sacrifice a creature: Destroy target creature with flying.",
			Purpose: game.Purpose{Answers: game.AnswerRemove | game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{4}{G}"), SacrificeACreature()),
			Targets: TargetCreature("target creature with flying", HasKeyword("flying")),
			Effect:  destroyFirstLegalTarget,
		}},
	})
}
