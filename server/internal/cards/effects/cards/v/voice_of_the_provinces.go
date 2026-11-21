package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Voice of the Provinces — {4}{W}{W} Creature — Angel 3/3 (Avacyn Restored, #5570):
//
//	"Flying
//	 When this creature enters, create a 1/1 white Human creature
//	 token."
//
// Flying and Clarion Cathars' enters token: one 1/1 white Human.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "81b15ed1-7069-4f8b-96b9-f6d67298afef",
		Name:            "Voice of the Provinces",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Voice of the Provinces — create a 1/1 white Human creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Human"), N: 1})),
		},
	})
}
