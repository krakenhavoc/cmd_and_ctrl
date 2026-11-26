package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Verdant Haven — Enchantment — Aura {2}{G} (Magic 2015, #5556):
//
//	"Enchant land
//	 When this Aura enters, you gain 2 life.
//	 Whenever enchanted land is tapped for mana, its controller adds an
//	 additional one mana of any color."
//
// Fertile Ground's mana trigger with an enters trigger that gains its
// controller 2 life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "09381494-393d-4ddf-b79f-7c0ae5dce912",
		Name:         "Verdant Haven",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Verdant Haven — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
		ManaTriggers: []game.ManaTrigger{
			effects.WheneverAttachedTapsForMana(
				"Verdant Haven — add an additional one mana of any color",
				"{W|U|B|R|G}"),
		},
	})
}
