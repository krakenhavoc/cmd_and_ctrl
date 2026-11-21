package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dragonlord's Servant — Creature — Goblin Shaman {1}{R}, 1/3
// (Foundations, #4465):
//
//	"Dragon spells you cast cost {1} less to cast."
//
// Dragonspeaker Shaman's discount at {1}. It reduces only the generic
// part of the cost (CR 601.2f), applies to any spell with the Dragon
// subtype that you cast, and reads "Dragon" through the card's text (CR
// 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "faf81ae0-0b35-45ef-a50d-6cecf0c0eeeb",
		Name:         "Dragonlord's Servant",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Dragon spells you cast cost {1} less to cast.",
				effects.YourSpell(), effects.SpellOfCreatureType("Dragon")),
		},
	})
}
