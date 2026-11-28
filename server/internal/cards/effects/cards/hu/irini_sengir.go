package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Irini Sengir — Legendary Creature — Vampire Dwarf {2}{B}{B}, 2/2 (Homelands,
// S82 OLD b826, #5635):
//
//	"Green enchantment spells and white enchantment spells cost {2} more to
//	 cast."
//
// A symmetrical tax (Thorn of Amethyst's shape, no "you cast"): every
// enchantment spell that is green or white, whoever casts it, costs {2} more
// (CR 601.2f). One modifier covers both colours, so a green-and-white
// enchantment spell pays the {2} once. Colour is read off the spell as it
// stands at announce, through the colour words of Irini's own text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2727a7ec-1cf7-4cd7-b5d4-955db940ddd9",
		Name:         "Irini Sengir",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsMore(2, "Green enchantment spells and white enchantment spells cost {2} more to cast.",
				func(q game.CostQuery) bool {
					return q.Card.IsEnchantment() &&
						(q.Card.HasColor(q.Source.TextColor("G")) || q.Card.HasColor(q.Source.TextColor("W")))
				}),
		},
	})
}
