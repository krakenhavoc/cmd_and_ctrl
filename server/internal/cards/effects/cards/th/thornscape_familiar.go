package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thornscape Familiar — Creature — Insect {1}{G}, 2/1 (Planeshift, S82 OLD b644, #5615):
//
//	"Red spells and white spells you cast cost {1} less to cast."
//
// Nightscape Familiar's two-colour discount (OldB644SpellsOfEitherColorCostLess): a
// spell of either colour, or both, costs {1} less of generic mana (CR 601.2f); the
// colours are words of this card's text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3f92c684-f2fd-4fe5-bfb5-e27061d2c45b",
		Name:         "Thornscape Familiar",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.OldB644SpellsOfEitherColorCostLess("Red spells and white spells you cast cost {1} less to cast.", "R", "W"),
		},
	})
}
