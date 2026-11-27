package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kagemaro's Clutch — Enchantment — Aura {3}{B} (Saviors of Kamigawa, #5601):
//
//	"Enchant creature
//	 Enchanted creature gets -X/-X, where X is the number of cards in your
//	 hand."
//
// A layer 7c modification (CR 613.4c) read live from the Aura's controller's
// hand, so it follows every draw and discard (DependsOnHandSize drops the
// layer cache when a hand moves).
//
// No simplification.
func init() {
	static := effects.PumpAttachedPer(-1, -1, func(g *game.Game, source *game.Card) int {
		return effects.CardsInHand(g, source.Controller)
	})
	static.Label = "Enchanted creature gets -X/-X, where X is the number of cards in your hand."
	static.DependsOnHandSize = true
	effects.Register(effects.Spec{
		OracleID:     "f83f6ba0-bb23-4c62-8a16-b17f5fbc0740",
		Name:         "Kagemaro's Clutch",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{static},
	})
}
