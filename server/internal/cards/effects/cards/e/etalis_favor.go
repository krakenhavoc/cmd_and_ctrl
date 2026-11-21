package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Etali's Favor — Enchantment — Aura {2}{R}:
//
//	"Enchant creature you control
//	 When this Aura enters, discover 3.
//	 Enchanted creature gets +1/+1 and has trample."
//
// Rancor's statics with an ETB discover. The discover is ADR 0099's
// (game/discover.go).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "df98b93b-0aa6-4bea-9d8d-798ebc795b0b",
		Name:         "Etali's Favor",
		Purpose:      game.Purpose{Discover: 3},
		Completeness: effects.CompletenessFull,
		Discovers:    true,
		Targets:      effects.EnchantCreature(effects.YouControl()),
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("trample"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Etali's Favor — discover 3", effects.DiscoverN(3)),
		},
	})
}
