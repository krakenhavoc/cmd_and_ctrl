package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mind Harness — Enchantment — Aura {U} (Mirage, S82 OLD batch 811, #5633):
//
//	"Enchant red or green creature
//	 Cumulative upkeep {1}
//	 You control enchanted creature."
//
// Mind Control's layer 2 control grant with Glacial Plating's cumulative
// upkeep. The enchant clause (red or green) is read as it stands every time
// state-based actions look (CR 704.5m).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bba30360-c54a-4906-8b2e-84bb1209c43d",
		Name:         "Mind Harness",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.Or(effects.OfColor("R"), effects.OfColor("G"))),
		Static:       []game.StaticAbility{effects.ControlAttachedBySource()},
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeep("Mind Harness — cumulative upkeep {1}", "{1}"),
		},
	})
}
