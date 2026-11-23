package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "0ccc75a0-1bc2-461f-ab6f-0f182328a863",
		Name:         "Inferno Fist",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Targets:      effects.EnchantCreature(effects.YouControl()),
		Static:       []game.StaticAbility{effects.PumpAttached(2, 0)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, Sacrifice this Aura: This Aura deals 2 damage to any target.",
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.SacrificeThis()),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(2),
		}},
	})
}
