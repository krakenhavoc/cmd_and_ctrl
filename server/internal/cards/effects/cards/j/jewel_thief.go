package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jewel Thief (Streets of New Capenna, #5491):
//
//	Vigilance, trample
//	When this creature enters, create a Treasure token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "470f0bbf-a0c9-4216-96f0-d400dac17aef",
		Name:            "Jewel Thief",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample"},
		Purpose:         game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Jewel Thief — create a Treasure token", effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
		},
	})
}
