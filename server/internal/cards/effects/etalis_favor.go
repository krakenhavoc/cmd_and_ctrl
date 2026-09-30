package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Etali's Favor — Enchantment — Aura {2}{R}:
//
//	"Enchant creature you control
//	 When this Aura enters, discover 3.
//	 Enchanted creature gets +1/+1 and has trample."
//
// Rancor's statics with an ETB discover. The discover is ADR 0099's
// (game/discover.go).
func init() {
	Register(Spec{
		OracleID:     "df98b93b-0aa6-4bea-9d8d-798ebc795b0b",
		Name:         "Etali's Favor",
		Completeness: CompletenessFull,
		Discovers:    true,
		Targets:      EnchantCreature(YouControl()),
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("trample"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Etali's Favor — discover 3", DiscoverN(3)),
		},
	})
}
