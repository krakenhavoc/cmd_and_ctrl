package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hermetic Study — Enchantment — Aura {1}{U} (Dominaria Remastered, #5948):
//
//	"Enchant creature
//	 Enchanted creature has "{T}: This creature deals 1 damage to any
//	 target.""
//
// The same ping bundle a Brimstone Mage's first level has, granted to the
// enchanted creature (ADR 0093); the creature deals the damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "792d7818-ee5c-4254-b969-7b50f475c629",
		Name:         "Hermetic Study",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Grants:       []effects.AbilityGrant{effects.RoeB1PingGrant(hermeticStudyGrant, 1)},
		Static:       []game.StaticAbility{effects.GrantAbilitiesToAttached(hermeticStudyGrant)},
	})
}

// hermeticStudyGrant names the granted ability's bundle: an on-disk identity.
const hermeticStudyGrant = "hermetic-study/tap-ping"
