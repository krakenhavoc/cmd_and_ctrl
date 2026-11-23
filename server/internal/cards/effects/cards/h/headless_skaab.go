package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Headless Skaab — Creature — Zombie Warrior {2}{U}, 3/6 (Dark Ascension, #5571):
//
//	"As an additional cost to cast this spell, exile a creature card
//	 from your graveyard.
//	 This creature enters tapped."
//
// Stitched Drake's additional cost (paid at CR 601.2h; with no creature
// card in the graveyard the spell can't be cast) and the shared enters-
// tapped replacement.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "203365d4-4793-47af-840e-418bd5569d50",
		Name:         "Headless Skaab",
		Completeness: effects.CompletenessFull,
		AdditionalCost: effects.ExileFromGraveyardCost(1, "a creature card", func(c game.Card) bool {
			return c.IsCreature()
		}),
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
	})
}
