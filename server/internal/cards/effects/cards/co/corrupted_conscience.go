package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Corrupted Conscience — Enchantment — Aura {3}{U}{U} (Mirrodin Besieged, S82
// OLD b468, #5576):
//
//	"Enchant creature
//	 You control enchanted creature.
//	 Enchanted creature has infect. (It deals damage to creatures in the form of
//	 -1/-1 counters and to players in the form of poison counters.)"
//
// Control Magic's layer-2 control change, which reverts by itself when the Aura
// leaves (CR 613.7 timestamps order two such effects), and Phyresis's
// layer-6 infect grant on the host, whoever controls it. Infect is the engine's
// damage tail (ADR 0056).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c645fab4-51ac-440c-b112-57777c8f4676",
		Name:         "Corrupted Conscience",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static: []game.StaticAbility{
			effects.ControlAttachedBySource(),
			effects.GrantToAttached("infect"),
		},
	})
}
