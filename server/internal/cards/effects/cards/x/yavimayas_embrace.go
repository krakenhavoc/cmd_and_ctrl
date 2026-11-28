package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Yavimaya's Embrace — Enchantment — Aura {5}{G}{U}{U} (Apocalypse, S82 OLD,
// #5614):
//
//	"Enchant creature
//	 You control enchanted creature.
//	 Enchanted creature gets +2/+2 and has trample."
//
// Mind Control's layer 2 control change, which reverts on its own when the Aura
// leaves, with Hypnotic Siren's pump and keyword on the same host: +2/+2 in
// layer 7c and trample in layer 6.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d5cb0370-a6f0-40c0-8b8f-553fca23778d",
		Name:         "Yavimaya's Embrace",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static: []game.StaticAbility{
			effects.ControlAttachedBySource(),
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("trample"),
		},
	})
}
