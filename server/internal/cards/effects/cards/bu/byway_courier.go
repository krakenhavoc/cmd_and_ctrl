package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Byway Courier — Creature — Human Scout {2}{G}, 3/2 (Foundations
// Jumpstart):
//
//	"When this creature dies, investigate."
//
// The dies trigger is Angelic Sleuth's investigate (CR 701.16a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e186cc6f-8c11-4475-a704-0b4c3a9eade9",
		Name:         "Byway Courier",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, Answers: game.AnswerValue},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Byway Courier — investigate", effects.Do(effects.CmdrB1Investigate())),
		},
	})
}
