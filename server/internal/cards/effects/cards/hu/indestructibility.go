package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Indestructibility — Enchantment — Aura {3}{W} (Magic 2014, S82 old sets):
//
//	"Enchant permanent
//	 Enchanted permanent has indestructible. (Effects that say "destroy" don't
//	 destroy that permanent. A creature with indestructible can't be destroyed
//	 by damage.)"
//
// A layer-6 grant of indestructible on whatever the Aura is attached to, any
// permanent. "Enchant permanent" is the widest spec, so an enchanted permanent
// that stops being a creature or a land keeps the Aura (Song of the Dryads'
// reasoning).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9aaa0a28-e342-48bd-8f59-0cd289491f98",
		Name:         "Indestructibility",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPermanent(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.GrantToAttached("indestructible")},
	})
}
