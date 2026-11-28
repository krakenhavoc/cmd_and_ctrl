package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gangrenous Zombies — Creature — Zombie {1}{B}{B}, 2/2 (Ice Age, S82 OLD, #5636):
//
//	"{T}, Sacrifice this creature: This creature deals 1 damage to each creature and
//	 each player. If you control a snow Swamp, this creature deals 2 damage to each
//	 creature and each player instead."
//
// Pestilence's sweep paid by the creature itself. The snow Swamp is checked as the ability
// resolves, and the damage is the Zombies' as it last existed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d0d844f2-6cf4-49f2-95db-64e932a15b17",
		Name:         "Gangrenous Zombies",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice this creature: This creature deals 1 damage to each creature and each player. If you control a snow Swamp, this creature deals 2 damage to each creature and each player instead.",
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, Partial: true}},
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Effect:  effects.S82OldB841GangrenousZombiesDamage,
		}},
	})
}
