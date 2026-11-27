package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arcane Melee — Enchantment {4}{U} (Commander 2013, S82 OLD b386):
//
//	"Instant and sorcery spells cost {2} less to cast."
//
// Mocking Sprite's reduction (generic only) without "you cast": it applies to
// every player's instant and sorcery spells.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9967b292-1239-4aed-b871-72e40d1f3e26",
		Name:         "Arcane Melee",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(2, "Instant and sorcery spells cost {2} less to cast.", effects.InstantOrSorcerySpell()),
		},
	})
}
