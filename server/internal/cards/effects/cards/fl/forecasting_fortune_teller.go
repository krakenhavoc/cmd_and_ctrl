package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Forecasting Fortune Teller — Creature — Human Advisor Ally {1}{U}, 1/3
// (Avatar: The Last Airbender, #4459):
//
//	"When this creature enters, create a Clue token. (It's an artifact
//	 with '{2}, Sacrifice this token: Draw a card.')"
//
// The catalog's Clue row. Creating a Clue this way is not investigating
// (the card's ruling), so nothing that watches for investigate fires.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a5a05edd-2a30-414b-b766-f0651d5783f5",
		Name:         "Forecasting Fortune Teller",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Forecasting Fortune Teller — create a Clue token",
				effects.Do(effects.CreateToken{Template: effects.ClueToken(), N: 1})),
		},
	})
}
