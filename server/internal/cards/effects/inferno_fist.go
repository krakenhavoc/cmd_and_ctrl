package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inferno Fist — Enchantment — Aura {1}{R} (Magic 2015, #5556):
//
//	"Enchant creature you control
//	 Enchanted creature gets +2/+0.
//	 {R}, Sacrifice this Aura: This Aura deals 2 damage to any target."
//
// Fire-Rim Form's +2/+0 on a host you control (EnchantCreature with
// YouControl), and Pyromania's sacrifice-for-damage row: the Aura is
// sacrificed as the cost and deals the damage as it last existed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0ccc75a0-1bc2-461f-ab6f-0f182328a863",
		Name:         "Inferno Fist",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Targets:      EnchantCreature(YouControl()),
		Static:       []game.StaticAbility{PumpAttached(2, 0)},
		Activated: []ActivatedAbility{{
			Label:   "{R}, Sacrifice this Aura: This Aura deals 2 damage to any target.",
			Cost:    Plus(ManaCost("{R}"), SacrificeThis()),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 2)),
			Effect:  sourceDealsDamageToEachLegalTarget(2),
		}},
	})
}
