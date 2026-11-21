package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Assassin's Ink — Instant {2}{B}{B} (Kamigawa: Neon Dynasty, #5493):
//
//	"This spell costs {1} less to cast if you control an artifact and {1} less
//	 to cast if you control an enchantment.
//	 Destroy target creature or planeswalker."
//
// Two self cost modifiers (CR 601.2f), one per sentence half, each a Cactarantula
// discount; both apply when both are met. A permanent that is both an artifact
// and an enchantment satisfies both.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a6d95d26-48d8-4a97-9a42-f492a61300c9",
		Name:         "Assassin's Ink",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(1, "This spell costs {1} less to cast if you control an artifact.", effects.YouControlA(effects.Artifact())),
			effects.CostsLess(1, "This spell costs {1} less to cast if you control an enchantment.", effects.YouControlA(effects.Enchantment())),
		},
		Targets:   effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: effects.DestroyTheTargetPermanent,
	})
}
