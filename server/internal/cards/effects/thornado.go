package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "7476a14a-611e-4f81-b603-fd600e88568f",
		Name:         "Thornado",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature with flying", HasKeyword("flying")),
		Activated:    []ActivatedAbility{Cycling("{1}{G}")},
		OnResolve:    destroyTheTargetPermanent,
	})
}
