package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Smite the Monstrous — Instant {3}{W} (Innistrad, #5572):
//
//	"Destroy target creature with power 4 or greater."
//
// Bring Down's target: power is judged at announce and again at
// resolution (CR 608.2b), so shrinking it below 4 in response counters
// the spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "649512ed-4908-450a-9652-b2c80df5e24d",
		Name:         "Smite the Monstrous",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature with power 4 or greater", effects.PowerGE(4)),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
