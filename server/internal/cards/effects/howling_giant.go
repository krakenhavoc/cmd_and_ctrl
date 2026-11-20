package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Howling Giant — Creature — Giant Druid {5}{G}{G}, 5/5 (Core Set 2020, #5518):
//
//	"Reach (This creature can block creatures with flying.)
//	 When this creature enters, create two 2/2 green Wolf creature tokens."
//
// Tovolar's Huntmaster's enters trigger with a printed reach.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c616242a-95f0-4cd8-99a1-9237cfea0e57",
		Name:            "Howling Giant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Purpose:         game.Purpose{Tokens: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Howling Giant — create two 2/2 green Wolf creature tokens",
				Do(CreateToken{Template: TokenCard("2/2 green Wolf"), N: 2})),
		},
	})
}
