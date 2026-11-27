package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Surging Might — Enchantment — Aura {2}{G} (Coldsnap, #5597):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2.
//	 Ripple 4 (...)"
//
// Spectral Steel's +2/+2 Aura with ripple (CR 702.60a) as a trigger of the
// spell. A free cast from the revealed cards is an Aura spell, so it picks its
// creature as it is cast.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "04647861-0023-432b-9b24-77264f11e470",
		Name:         "Surging Might",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(2, 2)},
		Triggered:    []game.TriggeredAbility{effects.S82OldB566Ripple("Surging Might", 4)},
	})
}
