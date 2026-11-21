package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// World-Weary — Enchantment — Aura {3}{B}{B} (Modern Horizons 2, #5501):
//
//	"Enchant creature
//	 Enchanted creature gets -4/-4.
//	 Basic landcycling {1}{B} ({1}{B}, Discard this card: Search your
//	 library for a basic land card, reveal it, put it into your hand,
//	 then shuffle.)"
//
// Failed Conversion's layer 7c -4/-4 on the host (PumpAttached) without
// its death trigger, and the cycling family's BasicLandcycling
// (CR 702.29e), which works only from hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "de6e0f80-3a9e-4499-aec1-2513a0264538",
		Name:         "World-Weary",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.PumpAttached(-4, -4)},
		Activated:    []effects.ActivatedAbility{effects.BasicLandcycling("{1}{B}")},
	})
}
