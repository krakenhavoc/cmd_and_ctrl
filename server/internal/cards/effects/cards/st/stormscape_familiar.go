package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stormscape Familiar — Creature — Bird {1}{U}, 1/1 (Planeshift, S82 OLD b643, #5615):
//
//	"Flying
//	 White spells and black spells you cast cost {1} less to cast."
//
// Nightscape Familiar's two-colour discount (OldB643SpellsOfEitherColorCostLess):
// a spell of either colour, or both, costs {1} less of generic mana (CR 601.2f);
// the colours are words of this card's text (CR 612.2). The keyword is printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a8f5a11f-3291-4ee9-a96d-4fa38af14932",
		Name:            "Stormscape Familiar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CostModifiers: []game.CostModifier{
			effects.OldB643SpellsOfEitherColorCostLess("White spells and black spells you cast cost {1} less to cast.", "W", "B"),
		},
	})
}
