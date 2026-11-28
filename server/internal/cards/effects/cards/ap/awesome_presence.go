package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Awesome Presence — Enchantment — Aura {U} (Alliances, #5634):
//
//	"Enchant creature
//	 Enchanted creature can't be blocked unless defending player pays {3} for
//	 each creature they control that's blocking it."
//
// The block tax of #6191, priced per (blocker, attacker) pair
// (EnchantedCantBeBlockedUnlessPay): every creature that would block the
// enchanted creature costs its controller {3} as blockers are declared
// (CR 509.1d, 509.1h). The Aura is the source, so the tax follows it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d6efff61-2087-4f48-a8a8-6c465190dad9",
		Name:         "Awesome Presence",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		BlockTaxes: []game.BlockTax{
			effects.EnchantedCantBeBlockedUnlessPay("{3}", "Awesome Presence — enchanted creature can't be blocked unless defending player pays {3} for each creature they control that's blocking it"),
		},
	})
}
