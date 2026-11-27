package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glacial Plating — Snow Enchantment — Aura {2}{W}{W} (Coldsnap, S82 OLD b559):
//
//	"Enchant creature
//	 Cumulative upkeep {S}
//	 Enchanted creature gets +3/+3 for each age counter on this Aura."
//
// {S} is paid with snow mana only, once per age counter (CR 107.4h); the pump
// reads the Aura's age counters on every recompute.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e9ee4257-5df8-4f4e-b821-b6313ded1802",
		Name:         "Glacial Plating",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttachedPer(3, 3, effects.OldB559AgeCountersOnSource)},
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeep("Glacial Plating — cumulative upkeep {S}", "{S}"),
		},
	})
}
