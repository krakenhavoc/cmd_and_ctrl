package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Market Festival — Enchantment — Aura {3}{G} (Journey into Nyx):
//
//	"Enchant land
//	 Whenever enchanted land is tapped for mana, its controller adds an
//	 additional two mana in any combination of colors."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "15c63f9b-d071-4522-b9f0-a1f441a61b81",
		Name:         "Market Festival",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		ManaTriggers: []game.ManaTrigger{
			effects.WheneverAttachedTapsForMana(
				"Market Festival — add an additional two mana in any combination of colors",
				effects.AnyCombinationOfColors(2)),
		},
	})
}
