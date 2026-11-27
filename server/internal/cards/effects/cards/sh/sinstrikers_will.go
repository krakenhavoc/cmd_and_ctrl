package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sinstriker's Will — Enchantment — Aura {3}{W} (Guildpact):
//
//	"Enchant creature
//	 Enchanted creature has "{T}: This creature deals damage equal to its
//	 power to target attacking or blocking creature.""
//
// Burning Anger's granted tap ability with an attacking or blocking creature
// as its target: the enchanted creature is the source, and its power is read
// as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "98e7b75b-f24d-4317-8b45-9397a4742081",
		Name:         "Sinstriker's Will",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: sinstrikersWillGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: This creature deals damage equal to its power to target attacking or blocking creature.",
				Cost:    effects.TapCost(),
				Targets: effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
				Effect:  effects.SourceDealsDamageEqualToItsPowerToTarget,
			}},
			Text: "{T}: This creature deals damage equal to its power to target attacking or blocking creature.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(sinstrikersWillGrant)},
	})
}

// sinstrikersWillGrant names the granted ability's bundle: an on-disk
// identity.
const sinstrikersWillGrant = "sinstrikers-will/tap-damage-equal-to-power"
