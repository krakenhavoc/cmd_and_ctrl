package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dwarven Scorcher — Creature — Dwarf {R}, 1/1 (Judgment, #5611):
//
//	"Sacrifice this creature: This creature deals 1 damage to target creature
//	 unless that creature's controller has this creature deal 2 damage to
//	 them."
//
// The creature's controller is asked as the ability resolves (the pay-unless
// prompt's "have the source deal damage to you" payment); a "no" lets it deal 1
// damage to the creature. The source is already gone (it was the cost), so both
// damages are dealt by its last-known self (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a896cedb-f10a-48ac-82e7-6ae1e8b5a439",
		Name:         "Dwarven Scorcher",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: This creature deals 1 damage to target creature unless that creature's controller has this creature deal 2 damage to them.",
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetCreature("target creature"),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.OldB623DwarvenScorcher,
		}},
	})
}
