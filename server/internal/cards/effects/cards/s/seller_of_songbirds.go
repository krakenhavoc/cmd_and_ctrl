package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seller of Songbirds — Creature — Human {2}{W}, 1/2 (Ravnica: Clue
// Edition):
//
//	"When this creature enters, create a 1/1 white Bird creature token with
//	 flying."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "87e4c6ff-f83f-412b-9d23-aac7a57ef6db",
		Name:         "Seller of Songbirds",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Seller of Songbirds — create a 1/1 white Bird creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Bird with flying"), N: 1})),
		},
	})
}
