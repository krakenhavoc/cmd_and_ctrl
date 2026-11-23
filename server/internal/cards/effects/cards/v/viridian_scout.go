package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Viridian Scout — Creature — Elf Warrior Scout {3}{G}, 1/2 (Fifth Dawn, #5604):
//
//	"{2}{G}, Sacrifice this creature: It deals 2 damage to target
//	 creature with flying."
//
// Barbarian Lunatic's sacrifice-and-shoot, limited to a flier (Centaur
// Archer's clause). The Scout is sacrificed as a cost, so the damage comes
// from it as it last existed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c5b4bc25-0f32-4e1d-b6fe-f122140f2207",
		Name:         "Viridian Scout",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}, Sacrifice this creature: It deals 2 damage to target creature with flying.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Cost:    effects.Plus(effects.ManaCost("{2}{G}"), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
			Effect:  effects.DealDamageToTheTarget(2),
		}},
	})
}
