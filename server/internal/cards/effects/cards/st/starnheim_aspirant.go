package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Starnheim Aspirant — Creature — Human Cleric {2}{W}, 2/2:
//
//	"Angel spells you cast cost {2} less to cast."
//
// Generic mana only, applied after any cost increase (CR 601.2f).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e7b92789-47a9-430a-ad57-be7b6c3d63cb",
		Name:         "Starnheim Aspirant",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(2, "Angel spells you cast cost {2} less to cast.",
				effects.YourSpell(), effects.SpellOfCreatureType("Angel")),
		},
	})
}
