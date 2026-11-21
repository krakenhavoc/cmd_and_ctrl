package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crippling Blight — Enchantment — Aura {B} (Magic 2013, #5568):
//
//	"Enchant creature
//	 Enchanted creature gets -1/-1 and can't block."
//
// Failed Conversion's shrink at -1/-1 with the host's can't-block
// restriction.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e365699a-bd77-473f-86b0-4e1ba0371cdc",
		Name:         "Crippling Blight",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.PumpAttached(-1, -1), effects.RestrictAttached(game.CantBlock)},
	})
}
