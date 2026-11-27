package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dawnray Archer — {2}{U} Creature — Human Archer 1/1 (Shards of Alara, S82
// OLD b515):
//
//	"Exalted (Whenever a creature you control attacks alone, that creature
//	 gets +1/+1 until end of turn.)
//	 {W}, {T}: This creature deals 1 damage to target attacking or blocking
//	 creature."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "19055292-1a79-413e-8056-7d4aec9839ed",
		Name:            "Dawnray Archer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"exalted"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}: This creature deals 1 damage to target attacking or blocking creature.",
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
			Targets: effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.DealDamageToTheTarget(1),
		}},
	})
}
