package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ancient Hellkite — Creature — Dragon {4}{R}{R}{R}, 6/6 (Magic 2011 / Game
// Night 2019, #5579):
//
//	"Flying
//	 {R}: This creature deals 1 damage to target creature defending player
//	 controls. Activate only if this creature is attacking."
//
// The restriction is SourceIsAttacking, checked at activation. The target
// is a creature controlled by the player the Hellkite is attacking, not by
// every opponent (S82OldB475ControlledByDefender, a new CombatRole on
// TargetSpec.CombatWithSource), judged again as the ability resolves: a
// Hellkite that has left combat leaves the ability with no legal target
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "de672a4b-07eb-4be8-a03f-23dfbb7ac0fa",
		Name:            "Ancient Hellkite",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:     "{R}: This creature deals 1 damage to target creature defending player controls. Activate only if this creature is attacking.",
			Cost:      effects.ManaCost("{R}"),
			Targets:   effects.S82OldB475ControlledByDefender(effects.TargetCreature("target creature defending player controls")),
			Purpose:   effects.ForTargets(effects.DamageToTarget(0, 1)),
			Condition: effects.SourceIsAttacking(),
			Effect:    effects.SourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
