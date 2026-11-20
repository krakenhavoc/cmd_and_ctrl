package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gallant Cavalry — Creature — Human Knight {3}{W}, 2/2
// (Core Set 2019, #5527):
//
//	"Vigilance (Attacking doesn't cause this creature to tap.)
//	When this creature enters, create a 2/2 white Knight creature token with vigilance."
//
// Valorous Steed's text: vigilance, and a 2/2 Knight token as it
// enters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b500c460-2814-47a4-b307-17ca519f3565",
		Name:            "Gallant Cavalry",
		Completeness:    CompletenessFull,
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Gallant Cavalry — create a 2/2 white Knight token with vigilance",
				Do(CreateToken{Template: TokenCard("2/2 white Knight with vigilance"), N: 1})),
		},
	})
}
