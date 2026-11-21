package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Divebomber Griffin — Creature — Griffin {3}{W}{W}, 3/2 (Ravnica: City
// of Guilds, #5600):
//
//	"Flying
//	 {T}, Sacrifice this creature: It deals 3 damage to target attacking
//	 or blocking creature."
//
// Fanatical Firebrand's tap-and-sacrifice row at 3 damage, aimed at a
// creature that is attacking or blocking as it is announced and still is
// at resolution (CR 608.2b). The Griffin is the source of the damage
// though it is already in the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "11e84e48-0f18-41b6-9f2a-30c64aecf9a7",
		Name:            "Divebomber Griffin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice this creature: It deals 3 damage to target attacking or blocking creature.",
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 3)),
			Effect:  effects.B33DamageChosenTargetFromSource(3),
		}},
	})
}
