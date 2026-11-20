package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Patagia Viper — Creature — Snake {3}{G}, 2/1 (Dissension, #5598):
//
//	"Flying
//	 When this creature enters, create two 1/1 green and blue Snake creature
//	 tokens.
//	 When this creature enters, sacrifice it unless {U} was spent to cast it."
//
// The Snakes are the token table's 1/1 green and blue Snake.
// The sacrifice is disSacrificeUnlessSpent: it reads what paid for the
// spell this permanent was (CR 400.7d), so a creature put onto the
// battlefield without being cast is sacrificed too. The auto-tapper is
// asked to spend the colour (WantsColors), an ordering and never a filter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b9985370-d908-431c-ad33-ec532fe07d1f",
		Name:            "Patagia Viper",
		Completeness:    CompletenessFull,
		ReadsManaSpent:  game.ReadsOwnPayment,
		WantsColors:     []string{"U"},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Patagia Viper — create two 1/1 green and blue Snakes", Do(CreateToken{Template: TokenCard("1/1 green and blue Snake"), N: 2})),
			disSacrificeUnlessSpent("Patagia Viper", "U"),
		},
	})
}
