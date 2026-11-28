package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fire Whip — Enchantment — Aura {1}{R} (Weatherlight):
//
//	"Enchant creature you control
//	 Enchanted creature has '{T}: This creature deals 1 damage to any target.'
//	 Sacrifice this Aura: It deals 1 damage to any target."
//
// The granted ability is the creature's own (it is the damage source, and its
// {T} obeys summoning sickness, CR 302.6). The Aura's sacrifice ability is the
// Aura's: the Aura deals the damage as it last existed (CR 608.2h).
//
// No simplification.
func init() {
	const grant = "fire-whip/ping"
	const text = "{T}: This creature deals 1 damage to any target."
	effects.Register(effects.Spec{
		OracleID:     "9d4372f0-eedf-4897-9f03-f96fb14ac4a6",
		Name:         "Fire Whip",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.YouControl()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   text,
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
				Cost:    effects.TapCost(),
				Targets: effects.TargetAny(),
				Effect:  effects.DealDamageToTheTarget(1),
			}},
			Text: text,
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(grant)},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this Aura: It deals 1 damage to any target.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetAny(),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
