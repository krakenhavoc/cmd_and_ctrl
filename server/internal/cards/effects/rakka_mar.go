package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rakka Mar — Legendary Creature — Human Shaman {2}{R}{R}, 2/2
// (Commander Legends):
//
//	"Haste
//	 {R}, {T}: Create a 3/1 red Elemental creature token with haste."
//
// A tap-and-pay token maker. The {T} is the symbol, so Rakka Mar cannot
// use it the turn he arrives unless he has haste, which he does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6d92f14b-470b-40bc-b294-f725d46f307e",
		Name:            "Rakka Mar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []ActivatedAbility{{
			Label:   "{R}, {T}: Create a 3/1 red Elemental creature token with haste.",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
			Cost:    Plus(ManaCost("{R}"), TapCost()),
			Effect:  Do(CreateToken{Template: TokenCard("3/1 red Elemental with haste"), N: 1}),
		}},
	})
}
