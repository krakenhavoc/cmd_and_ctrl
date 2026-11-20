package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "de6e0f80-3a9e-4499-aec1-2513a0264538",
		Name:         "World-Weary",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{PumpAttached(-4, -4)},
		Activated:    []ActivatedAbility{BasicLandcycling("{1}{B}")},
	})
}
