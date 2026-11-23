package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Knight of the New Coalition — Creature — Human Knight {3}{W}, 2/2
// (March of the Machine):
//
//	"Vigilance
//	 When this creature enters, create a 2/2 white and blue Knight
//	 creature token with vigilance."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "30ab812f-a055-4c68-ad81-9dd23e38827f",
		Name:            "Knight of the New Coalition",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Knight of the New Coalition — create a 2/2 white and blue Knight with vigilance",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 white and blue Knight with vigilance"), N: 1})),
		},
	})
}
