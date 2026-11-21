package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Geist of Saint Thalia — Legendary Creature — Spirit Cleric {1}{U}, 1/2:
//
//	"Flying
//	 Noncreature spells you cast cost {1} less to cast."
//
// Goblin Electromancer's shape with NoncreatureSpell in place of
// InstantOrSorcerySpell. The reduction spends generic mana only, so it
// never eats a coloured pip.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ef32a4a9-14e2-4738-b4c2-53ce5e1d2a53",
		Name:            "Geist of Saint Thalia",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Noncreature spells you cast cost {1} less to cast.",
				effects.YourSpell(), effects.NoncreatureSpell()),
		},
	})
}
