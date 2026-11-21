package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scourgemark — Enchantment — Aura {1}{B} (Theros, #5561):
//
//	"Enchant creature
//	 When this Aura enters, draw a card.
//	 Enchanted creature gets +1/+0."
//
// Chosen by Heliod's shape: a static pump and an enters trigger that
// draws.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "66efb37f-9eab-4ff4-a1d4-65fe9ca7332b",
		Name:         "Scourgemark",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Draws: 1},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 0),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Scourgemark — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
