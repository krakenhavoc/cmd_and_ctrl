package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fragmentize — Sorcery {W} (Kaladesh, #5539):
//
//	"Destroy target artifact or enchantment with mana value 4 or less."
//
// Naturalize with a mana value bound on the target, judged at announce
// and again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f025c1da-734e-4cbb-99d8-9d700d6edbde",
		Name:         "Fragmentize",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact or enchantment with mana value 4 or less", effects.Or(effects.Artifact(), effects.Enchantment()), effects.ManaValueLE(4)),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
