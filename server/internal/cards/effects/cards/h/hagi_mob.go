package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hagi Mob — Creature — Troll Berserker {4}{R}, 5/4 (Kaldheim, #5505):
//
//	"Boast — {1}{R}: This creature deals 1 damage to any target.
//	 (Activate only if this creature attacked this turn and only once
//	 each turn.)"
//
// The boast constructor (boast.go, CR 702.142) carries the "attacked
// this turn, once each turn" gate; the body is Anaba Shaman's ping.
//
// No simplification.
func init() {
	ab := effects.BoastTargeting("{1}{R}: This creature deals 1 damage to any target.",
		effects.ManaCost("{1}{R}"), effects.TargetAny(), effects.DealDamageToTheTarget(1))
	ab.Purpose = effects.ForTargets(effects.DamageToTarget(0, 1))
	effects.Register(effects.Spec{
		OracleID:     "c19aaf86-b098-4e7d-af50-9e02ae3cda55",
		Name:         "Hagi Mob",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{ab},
	})
}
