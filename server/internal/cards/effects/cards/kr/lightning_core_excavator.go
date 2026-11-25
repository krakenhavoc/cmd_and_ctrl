package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lightning-Core Excavator — Artifact Creature — Golem {1}, 0/3 (Jumpstart,
// #5510):
//
//	"{5}, {T}, Sacrifice this creature: It deals 3 damage to any target."
//
// Pyrite Spellbomb's sacrifice-for-damage. The sacrifice and the tap are
// the cost, so the Excavator is gone before the ability resolves and deals
// its damage as it last existed (CR 608.2h); summoning sickness gates the
// {T} (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cdaf3711-5bec-41bf-b333-c5f34ccb12ef",
		Name:         "Lightning-Core Excavator",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}, {T}, Sacrifice this creature: It deals 3 damage to any target.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 3)),
			Cost:    effects.Plus(effects.ManaCost("{5}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetAny(),
			Effect:  effects.DealDamageToTheTarget(3),
		}},
	})
}
