package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lightning Prowess — {2}{R} Aura: "Enchant creature. Enchanted creature
// has haste and "{T}: This creature deals 1 damage to any target.""
//
// No simplification.
func init() {
	const grant = "lightning-prowess/ping"
	effects.Register(effects.Spec{
		OracleID:     "85f40ced-a065-470a-a686-faaca17fa40c",
		Name:         "Lightning Prowess",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants:       []effects.AbilityGrant{effects.RoeB1PingGrant(grant, 1)},
		Static: []game.StaticAbility{
			effects.GrantToAttached("haste"),
			effects.GrantAbilitiesToAttached(grant),
		},
	})
}
