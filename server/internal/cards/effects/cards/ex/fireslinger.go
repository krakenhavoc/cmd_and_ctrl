package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fireslinger — Creature — Human Wizard {1}{R}, 1/1 (Jumpstart 2022,
// #5948):
//
//	"{T}: This creature deals 1 damage to any target and 1 damage to you."
//
// Goblin Artillery's tap ability at 1 and 1 (Forge Devil's body,
// TmtB6DamageAnyTargetAndYou): the damage to the target and to its controller
// is dealt at once, as one damage instance. If the target is gone, the ability
// does nothing, the damage to you included (CR 608.2b). A tap ability, so
// summoning sickness applies (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0e5a5cee-d068-42cf-bf58-8e1360720be0",
		Name:         "Fireslinger",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to any target and 1 damage to you.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.TmtB6DamageAnyTargetAndYou(1, 1),
		}},
	})
}
