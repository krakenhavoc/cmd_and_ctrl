package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Luminous Angel — Creature — Angel {4}{W}{W}{W}, 4/4 (Mirrodin, #5606):
//
//	"Flying
//	 At the beginning of your upkeep, you may create a 1/1 white Spirit
//	 creature token with flying."
//
// Printed flying, and Creakwood Liege's optional upkeep token: the
// "you may" is asked as the trigger resolves (CR 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8d7cf56c-94fd-47d8-9e0f-cd3163688983",
		Name:            "Luminous Angel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.AtYourUpkeep("Luminous Angel — create a 1/1 white Spirit creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Spirit with flying"), N: 1})),
				"Luminous Angel — create a 1/1 white Spirit creature token with flying?"),
		},
	})
}
