package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ruby Leech — Creature — Leech {1}{R}, 2/2 (Invasion, S82 #5616):
//
//	"First strike
//	 Red spells you cast cost {R} more to cast."
//
// Sapphire Leech's static cost increase with a red unit on your own red spells.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ec6195a0-2701-40a4-882f-eb563668791a",
		Name:            "Ruby Leech",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		CostModifiers: []game.CostModifier{
			effects.CostsMoreColored("{R}", "Red spells you cast cost {R} more to cast.", effects.YourSpell(), effects.ColoredSpell("R")),
		},
	})
}
