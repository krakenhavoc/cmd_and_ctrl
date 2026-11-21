package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Assembly Specialist — Artifact Creature — Borg {1}{W}, 1/1
// (Star Trek, #4973):
//
//	"When this creature enters, create a 1/1 colorless Borg artifact
//	 creature token."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "deb26a1b-2b83-4da2-a2e5-31c17a57eb1e",
		Name:         "Assembly Specialist",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Assembly Specialist — create a 1/1 colorless Borg artifact creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Borg artifact"), N: 1})),
		},
	})
}
