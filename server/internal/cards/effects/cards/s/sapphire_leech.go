package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sapphire Leech — Creature — Leech {1}{U}, 2/2 (Invasion, S82 #5616):
//
//	"Flying
//	 Blue spells you cast cost {U} more to cast."
//
// A static cost increase on your own blue spells with a coloured unit
// (CostsMoreColored): the {U} is added to the total cost (CR 601.2f), and the
// Leech's own spell is blue, so it is not taxed (it is not on the battlefield
// as it is cast).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "500aff17-0d69-46ac-a4f9-2fc6db061ce7",
		Name:            "Sapphire Leech",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CostModifiers: []game.CostModifier{
			effects.CostsMoreColored("{U}", "Blue spells you cast cost {U} more to cast.", effects.YourSpell(), effects.ColoredSpell("U")),
		},
	})
}
