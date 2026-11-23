package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ballyrush Banneret — Creature — Kithkin Soldier {1}{W}, 2/1
// (Foundations, #4465):
//
//	"Kithkin spells and Soldier spells you cast cost {1} less to cast."
//
// Gagh Slinger's discount with Kithkin and Soldier in place of Klingon
// and Warrior: one cost modifier (CR 601.2f) whose predicate takes a spell
// with either type, so a Kithkin Soldier spell is {1} cheaper, not {2}. It
// spends against generic mana only. A changeling spell and a kindred spell
// of either type count; the types are words of the Banneret's text (CR
// 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0ad27d0e-b6c3-4c3f-a389-285d889b6a65",
		Name:         "Ballyrush Banneret",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Kithkin spells and Soldier spells you cast cost {1} less to cast.",
				effects.YourSpell(), effects.SpellOfEitherCreatureType("Kithkin", "Soldier")),
		},
	})
}
