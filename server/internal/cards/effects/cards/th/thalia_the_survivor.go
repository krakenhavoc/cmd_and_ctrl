package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thalia, the Survivor — Legendary Creature — Human Soldier {3}{W}, 3/4:
//
//	"Lifelink
//	 Noncreature spells your opponents cast cost {1} more to cast."
//
// Thalia, Guardian of Thraben's tax narrowed to opponents' spells.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "55ef129e-698e-424f-be3c-3fbba6c2cc3e",
		Name:            "Thalia, the Survivor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		CostModifiers: []game.CostModifier{
			effects.CostsMore(1, "Noncreature spells your opponents cast cost {1} more to cast.",
				effects.OpponentsSpell(), effects.NoncreatureSpell()),
		},
	})
}
