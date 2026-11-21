package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Centaur Archer — Creature — Centaur Archer {1}{R}{G}, 3/2 (Ice Age,
// #5636):
//
//	"{T}: This creature deals 1 damage to target creature with flying."
//
// The target clause is a creature with flying, read through the layers
// at announce and again as it resolves, so a creature that lost flying
// in response is no longer a legal target (CR 608.2b). The damage is the
// shared DealDamageToTheTarget body, dealt by the Archer. The tap means
// summoning sickness applies (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d48eb6ee-d4ca-4bc7-8d7b-0e6261b88ae5",
		Name:         "Centaur Archer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to target creature with flying.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
			Effect:  effects.DealDamageToTheTarget(1),
		}},
	})
}
