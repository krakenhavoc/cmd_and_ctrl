package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scavenged Weaponry — Enchantment — Aura {2}{B} (Invasion, S82 older sets batch 661, #5616):
//
//	"Enchant creature
//	 When this Aura enters, draw a card.
//	 Enchanted creature gets +1/+1."
//
// A plain enters trigger and a layer 7c boost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2d3010d5-5c21-4342-ad79-a737b6731230",
		Name:         "Scavenged Weaponry",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Draws: 1},
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Scavenged Weaponry — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
