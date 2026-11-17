package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Knight of the New Coalition — Creature — Human Knight {3}{W}, 2/2
// (March of the Machine):
//
//	"Vigilance
//	 When this creature enters, create a 2/2 white and blue Knight
//	 creature token with vigilance."
func init() {
	Register(Spec{
		OracleID:        "30ab812f-a055-4c68-ad81-9dd23e38827f",
		Name:            "Knight of the New Coalition",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Knight of the New Coalition — create a 2/2 white and blue Knight with vigilance",
				Do(CreateToken{Template: TokenCard("2/2 white and blue Knight with vigilance"), N: 1})),
		},
	})
}
