package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Awaken the Ancient — Enchantment — Aura {1}{R}{R}{R} (Magic 2014, S82 old
// sets):
//
//	"Enchant Mountain
//	 Enchanted Mountain is a 7/7 red Giant creature with haste. It's still a
//	 land."
//
// Vastwood Zendikon's animation as a static ability of the Aura: the Creature
// type and the Giant subtype in layer 4, red in layer 5, base 7/7 in layer 7b,
// and a layer-6 haste grant, all while the Aura is on it. "Enchant Mountain" is
// the enchant-land spec narrowed to the Mountain subtype, the one the CR 704.5m
// legality check reruns, so a land that stops being a Mountain takes the Aura
// with it. A land that is already a creature is simply a 7/7.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0fd27f59-36ec-4235-8e1f-4a4a14972f4a",
		Name:         "Awaken the Ancient",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(effects.HasSubtype("Mountain")),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: append(effects.S82OldB100AnimateAttachedLand("R", "Giant", 7, 7),
			effects.GrantToAttached("haste")),
	})
}
