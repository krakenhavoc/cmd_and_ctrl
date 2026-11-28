package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hero's Resolve — Enchantment — Aura {1}{W} (Tempest, S82 OLD, #5628):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+5."
//
// A layer 7c modification (CR 613.4c) read off the attachment each pass, so it
// follows the Aura when it is moved and ends with it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1854a99d-f8c7-45b3-83a2-98ae1c5b5b09",
		Name:         "Hero's Resolve",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(1, 5)},
	})
}
