package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skyway Sniper — Creature — Elf Archer {G}, 1/2 (Core Set 2021, #5511):
//
//	"Reach
//	 {2}{G}: This creature deals 1 damage to target creature with flying."
//
// Centaur Archer's shot at a flyer (DealDamageToTheTarget) for a mana
// cost rather than a tap; the flying clause is re-checked as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8d907825-25f9-4b25-9fac-e0a207ed1797",
		Name:            "Skyway Sniper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}: This creature deals 1 damage to target creature with flying.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Cost:    effects.ManaCost("{2}{G}"),
			Targets: effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
			Effect:  effects.DealDamageToTheTarget(1),
		}},
	})
}
