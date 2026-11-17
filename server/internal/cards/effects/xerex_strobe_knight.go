package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Xerex Strobe-Knight — Creature — Human Knight {2}{U}, 2/2 (March of the
// Machine):
//
//	"Flying, vigilance
//	 {T}: Create a 2/2 white and blue Knight creature token with
//	 vigilance. Activate only if you've cast two or more spells this turn."
//
// The gate is the per-turn cast tally (copies are not cast, CR 707.10).
func init() {
	Register(Spec{
		OracleID:        "2e150f97-5688-4559-94b8-86c33fda28a0",
		Name:            "Xerex Strobe-Knight",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Activated: []ActivatedAbility{{
			Label:     "{T}: Create a 2/2 white and blue Knight creature token with vigilance. Activate only if you've cast two or more spells this turn.",
			Purpose:   game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:      TapCost(),
			Condition: s82Mom09CastTwoOrMoreSpellsThisTurn,
			Effect:    Do(CreateToken{Template: TokenCard("2/2 white and blue Knight with vigilance"), N: 1}),
		}},
	})
}
