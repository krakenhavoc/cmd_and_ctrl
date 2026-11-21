package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vryn Wingmare — Creature — Pegasus {2}{W}, 2/1 (Magic Origins, #5549):
//
//	"Flying
//	 Noncreature spells cost {1} more to cast."
//
// Flying, and Thalia's tax on every player's noncreature spells.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c7c241f2-d3cc-4032-a9da-149de73c2539",
		Name:            "Vryn Wingmare",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CostModifiers: []game.CostModifier{
			effects.CostsMore(1, "Noncreature spells cost {1} more to cast.", effects.NoncreatureSpell()),
		},
	})
}
