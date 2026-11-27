package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Raka Sanctuary — Enchantment {2}{R} (Apocalypse, #5614):
//
//	"At the beginning of your upkeep, if you control a white or blue
//	 permanent, this enchantment deals 1 damage to target creature. If you
//	 control a white permanent and a blue permanent, this enchantment deals 3
//	 damage instead."
//
// One triggered ability with an intervening if (CR 603.4), checked as it
// triggers and again on resolution. The target is chosen as it goes on the
// stack; the damage is 3 when both colours are controlled as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0b085c90-95df-4823-ba67-9eb398b1ec94",
		Name:         "Raka Sanctuary",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.S82OldB633RakaSanctuary("Raka Sanctuary — 1 damage to target creature (3 with a white and a blue permanent)"),
		},
	})
}
