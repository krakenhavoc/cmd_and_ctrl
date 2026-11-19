package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seller of Songbirds — Creature — Human {2}{W}, 1/2 (Ravnica: Clue
// Edition):
//
//	"When this creature enters, create a 1/1 white Bird creature token with
//	 flying."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "87e4c6ff-f83f-412b-9d23-aac7a57ef6db",
		Name:         "Seller of Songbirds",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Seller of Songbirds — create a 1/1 white Bird creature token with flying",
				Do(CreateToken{Template: TokenCard("1/1 white Bird with flying"), N: 1})),
		},
	})
}
