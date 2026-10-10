package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vitu-Ghazi Guildmage — Creature {G}{W} 2/2, Dryad Shaman:
//
//	"{4}{G}{W}: Create a 3/3 green Centaur creature token.
//	 {2}{G}{W}: Populate. (Create a token that's a copy of a creature
//	 token you control.)"
//
// Two mana-only activated abilities, neither with a tap symbol, so
// the Guildmage can be activated the turn it arrives and as often as
// the mana allows. The second is Populate (populate.go); it can copy
// the Centaur the first one makes, which is the point of the card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "da6f79ec-5f7e-4099-89c5-6d2bfc14d957",
		Name:         "Vitu-Ghazi Guildmage",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{4}{G}{W}: Create a 3/3 green Centaur creature token.",
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
				Cost:    ManaCost("{4}{G}{W}"),
				Effect:  Do(CreateToken{Template: TokenCard("3/3 green Centaur"), N: 1}),
			},
			{
				Label:   "{2}{G}{W}: Populate.",
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
				Cost:    ManaCost("{2}{G}{W}"),
				Effect:  Do(Populate{}),
			},
		},
	})
}
