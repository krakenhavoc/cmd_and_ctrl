package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thunderscape Familiar — Creature — Kavu {1}{R}, 1/1 (Planeshift, S82 OLD b643, #5615):
//
//	"First strike
//	 Black spells and green spells you cast cost {1} less to cast."
//
// Nightscape Familiar's two-colour discount (OldB643SpellsOfEitherColorCostLess):
// a spell of either colour, or both, costs {1} less of generic mana (CR 601.2f);
// the colours are words of this card's text (CR 612.2). The keyword is printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "40f6ea6f-71e1-46bd-8d9a-ccb5f437892c",
		Name:            "Thunderscape Familiar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		CostModifiers: []game.CostModifier{
			effects.OldB643SpellsOfEitherColorCostLess("Black spells and green spells you cast cost {1} less to cast.", "B", "G"),
		},
	})
}
