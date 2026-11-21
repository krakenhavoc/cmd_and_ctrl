package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thornweald Archer — Creature — Elf Archer {1}{G}, 2/1 (Foundations,
// #4465):
//
//	"Reach
//	 Deathtouch"
//
// Two printed keywords the engine enforces from PrintedKeywords: it can
// block a flier, and any damage it deals to a creature is lethal.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b5ba87f5-3589-4a8b-ad65-2ae2523e98e8",
		Name:            "Thornweald Archer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "deathtouch"},
	})
}
