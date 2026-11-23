package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phyresis — Enchantment — Aura for {1}{B}:
//
//	"Enchant creature
//	 Enchanted creature has infect. (It deals damage to creatures in the
//	 form of -1/-1 counters and to players in the form of poison
//	 counters.)"
//
// An ordinary keyword grant to the attached creature: the static
// re-reads the attachment on every layer pass, so the creature has
// infect for exactly as long as the Aura is on it, and the engine's
// infect rules (CR 702.90) do the rest.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "958628af-243c-4eb0-8620-5d483b6da768",
		Name:         "Phyresis",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static: []game.StaticAbility{
			effects.GrantToAttached("infect"),
		},
	})
}
