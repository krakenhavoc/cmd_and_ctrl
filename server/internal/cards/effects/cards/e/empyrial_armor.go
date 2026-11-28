package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Empyrial Armor — Enchantment — Aura {1}{W}{W} (Weatherlight):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 for each card in your hand."
//
// "Your" is the Aura's controller, not the enchanted creature's: the count is
// re-read on every recompute, so the bonus follows the hand as it fills and
// empties.
//
// No simplification.
func init() {
	pump := effects.PumpAttachedPer(1, 1, func(g *game.Game, source *game.Card) int {
		return effects.CardsInHand(g, source.Controller)
	})
	pump.DependsOnHandSize = true
	effects.Register(effects.Spec{
		OracleID:     "ec9e8611-e785-4d5d-b6e1-bf63daf408df",
		Name:         "Empyrial Armor",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{pump},
	})
}
