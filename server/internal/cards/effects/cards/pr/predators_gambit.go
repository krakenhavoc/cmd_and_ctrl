package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Predator's Gambit — {B} Aura: "Enchant creature. Enchanted creature gets
// +2/+1. Enchanted creature has intimidate as long as its controller
// controls no other creatures."
//
// The intimidate is a conditional grant re-read every layer pass.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4a1f94c0-6858-4296-a4a6-27d2201ed7f4",
		Name:         "Predator's Gambit",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 1),
			effects.GrantToAttachedWhile(effects.HostControllerControlsNoOtherCreatures, "intimidate"),
		},
	})
}
