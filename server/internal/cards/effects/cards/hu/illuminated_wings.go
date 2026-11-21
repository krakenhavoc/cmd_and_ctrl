package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Illuminated Wings — Enchantment — Aura {1}{U} (Urza's Destiny, #5621):
//
//	"Enchant creature
//	 Enchanted creature has flying.
//	 {2}, Sacrifice this Aura: Draw a card."
//
// Capashen Standard's sacrifice row on a flying Aura: sacrificing it is
// a cost, so the creature loses flying at once, before the draw
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7a22389a-34e7-4726-a551-f6fbc225cefe",
		Name:         "Illuminated Wings",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.GrantToAttached("flying")},
		Activated:    []effects.ActivatedAbility{effects.UdsB3SacrificeThisDrawRow("Aura")},
	})
}
