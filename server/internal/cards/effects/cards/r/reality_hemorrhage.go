package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reality Hemorrhage — Instant {1}{R} (Oath of the Gatewatch, #5545):
//
//	"Devoid (This card has no color.)
//	 Reality Hemorrhage deals 2 damage to any target."
//
// Devoid makes it colourless (CR 702.114a). Shock's body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "21465dcd-c93a-4a2c-ae10-e22ebd5cae2f",
		Name:            "Reality Hemorrhage",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Targets:         effects.TargetAny(),
		Purpose:         game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 2)), Answers: game.AnswerRemove},
		OnResolve:       effects.DamageToFirstTargetStep(2),
	})
}
