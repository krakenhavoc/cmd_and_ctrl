package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mark of the Vampire — Enchantment — Aura {3}{B} (Magic 2013, #5568):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 and has lifelink."
//
// The attachment statics: a layer 7c pump and a layer 6 lifelink grant
// on the host.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "af942d30-a191-4306-846d-6c26755ca3e6",
		Name:         "Mark of the Vampire",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(2, 2), effects.GrantToAttached("lifelink")},
	})
}
