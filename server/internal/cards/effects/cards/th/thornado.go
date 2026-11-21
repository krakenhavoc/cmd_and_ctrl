package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thornado — Instant {2}{G}:
//
//	"Destroy target creature with flying.
//	 Cycling {1}{G}"
//
// The flying clause is read through the layered view at announce and
// again at resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7476a14a-611e-4f81-b603-fd600e88568f",
		Name:         "Thornado",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
		Activated:    []effects.ActivatedAbility{effects.Cycling("{1}{G}")},
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
