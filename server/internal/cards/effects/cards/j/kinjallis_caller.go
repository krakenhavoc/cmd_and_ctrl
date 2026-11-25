package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kinjalli's Caller — Creature — Human Cleric {W}, 0/3 (Ixalan, #5532):
//
//	"Dinosaur spells you cast cost {1} less to cast."
//
// Ballyrush Banneret's discount with one type: a cost modifier (CR 601.2f)
// on spells you cast with the Dinosaur subtype, read off the spell on the
// stack so a changeling counts. It spends against generic mana only and
// never below the coloured cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bba6075b-8bbb-4496-80a9-bfa7f003d1e8",
		Name:         "Kinjalli's Caller",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Dinosaur spells you cast cost {1} less to cast.",
				effects.YourSpell(), effects.SpellOfCreatureType("Dinosaur")),
		},
	})
}
