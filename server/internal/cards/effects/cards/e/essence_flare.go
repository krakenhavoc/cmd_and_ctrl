package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Essence Flare — Enchantment — Aura {U} (Ice Age, S82 OLD #5636):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+0.
//	 At the beginning of the upkeep of enchanted creature's controller, put a
//	 -0/-1 counter on that creature."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9bdc79c9-c8b0-4db1-89d9-0ca7920b6576",
		Name:         "Essence Flare",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(2, 0)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, effects.UpkeepOfEnchantedCreaturesController,
				"Essence Flare — put a -0/-1 counter on enchanted creature", effects.EssenceFlareCounterB829),
		},
	})
}
