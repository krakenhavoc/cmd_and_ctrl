package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Xerex Strobe-Knight — Creature — Human Knight {2}{U}, 2/2 (March of the
// Machine):
//
//	"Flying, vigilance
//	 {T}: Create a 2/2 white and blue Knight creature token with
//	 vigilance. Activate only if you've cast two or more spells this turn."
//
// The gate is the per-turn cast tally (copies are not cast, CR 707.10).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2e150f97-5688-4559-94b8-86c33fda28a0",
		Name:            "Xerex Strobe-Knight",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:     "{T}: Create a 2/2 white and blue Knight creature token with vigilance. Activate only if you've cast two or more spells this turn.",
			Purpose:   game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:      effects.TapCost(),
			Condition: effects.S82Mom09CastTwoOrMoreSpellsThisTurn,
			Effect:    effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 white and blue Knight with vigilance"), N: 1}),
		}},
	})
}
