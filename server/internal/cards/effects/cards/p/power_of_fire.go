package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Power of Fire — Enchantment — Aura {1}{R} (Conspiracy, S82 OLD batch 362,
// #5557):
//
//	"Enchant creature
//	 Enchanted creature has "{T}: This creature deals 1 damage to any
//	 target.""
//
// The ping bundle Hermetic Study grants (RoeB1PingGrant), granted to the
// enchanted creature (ADR 0093); the creature deals the damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "df4903cb-50fa-4620-8470-6a6a398fc1ef",
		Name:         "Power of Fire",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Grants:       []effects.AbilityGrant{effects.RoeB1PingGrant(powerOfFireGrant, 1)},
		Static:       []game.StaticAbility{effects.GrantAbilitiesToAttached(powerOfFireGrant)},
	})
}

// powerOfFireGrant names the granted ability's bundle: an on-disk identity.
const powerOfFireGrant = "power-of-fire/tap-ping"
