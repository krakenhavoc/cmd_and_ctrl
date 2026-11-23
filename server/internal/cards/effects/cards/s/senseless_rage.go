package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Senseless Rage — Enchantment — Aura {1}{R} (Shadows over Innistrad, #5543):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2.
//	 Madness {1}{R} (If you discard this card, discard it into exile.
//	 When you do, cast it for its madness cost or put it into your
//	 graveyard.)"
//
// Spectral Steel's +2/+2 Aura (PumpAttached) without the graveyard
// ability. Madness is grown by the engine from the one string (CR 702.35a,
// #657).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5036fe35-43c6-45af-99d5-7b79ebcfc406",
		Name:         "Senseless Rage",
		Completeness: effects.CompletenessFull,
		Madness:      "{1}{R}",
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(2, 2)},
	})
}
