package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Krovikan Whispers — Enchantment — Aura {3}{U} (Coldsnap, S82 OLD b558):
//
//	"Enchant creature
//	 Cumulative upkeep {U} or {B}
//	 You control enchanted creature.
//	 When this Aura is put into a graveyard from the battlefield, you lose 2
//	 life for each age counter on it."
//
// Mind Control's control-changing static, a hybrid cumulative upkeep (one
// {U/B} symbol per age counter) and a graveyard trigger that reads the Aura's
// last-known age counters, whether it was sacrificed or fell off with the
// creature (CR 603.10a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4baf3db8-6f35-4101-9b21-15f4384c249a",
		Name:         "Krovikan Whispers",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static:       []game.StaticAbility{effects.ControlAttachedBySource()},
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeep("Krovikan Whispers — cumulative upkeep {U} or {B}", "{U/B}"),
			effects.WhenThisDies("Krovikan Whispers — you lose 2 life for each age counter on it",
				func(g *game.Game, item *game.StackItem) error {
					n := effects.AgeCountersAsItLeft(g, item)
					if n <= 0 {
						return nil
					}
					return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -2*n)
				}),
		},
	})
}
