package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Collar the Culprit — Instant {3}{W} (Guilds of Ravnica, #5525):
//
//	"Destroy target creature with toughness 4 or greater."
//
// Sungold Barrage's body. Toughness is read as the spell is cast and again
// as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "adc03c37-33c7-4b92-b9fd-4608e6c372c4",
		Name:         "Collar the Culprit",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature with toughness 4 or greater", effects.ToughnessGE(4)),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve:    effects.S82MidDestroyTheTarget,
	})
}
