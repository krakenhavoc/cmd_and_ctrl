package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sunscape Familiar — Creature — Wall {1}{W}, 0/3 (Planeshift, S82 OLD b643, #5615):
//
//	"Defender (This creature can't attack.)
//	 Green spells and blue spells you cast cost {1} less to cast."
//
// Nightscape Familiar's two-colour discount (OldB643SpellsOfEitherColorCostLess):
// a spell of either colour, or both, costs {1} less of generic mana (CR 601.2f);
// the colours are words of this card's text (CR 612.2). The keyword is printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "615603bb-3899-4419-9444-e4b64c94a438",
		Name:            "Sunscape Familiar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		CostModifiers: []game.CostModifier{
			effects.OldB643SpellsOfEitherColorCostLess("Green spells and blue spells you cast cost {1} less to cast.", "G", "U"),
		},
	})
}
